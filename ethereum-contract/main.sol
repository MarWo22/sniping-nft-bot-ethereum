pragma solidity ^0.8.7;

import "interfaces.sol";

contract EncodedNFTDelegatorEfficient
{ 
    struct EncodedAditionalRecipientOpenSea {
        bytes20 recipient;
        bytes32 amount;
    }

    struct EncodedParametersOpenSea {
        bytes1 considerationIdentifier; // 0x44 //
        bytes1 offerAmount; // 0x104
        bytes1 basicOrderType; // 0x124
        bytes1 totalOriginalAdditionalRecipients; // 0x204
        bytes2 offerIdentifier; // 0xe4
        bytes4 startTime; // 0x144
        bytes4 endTime; // 0x164
        bytes20 considerationToken; // 0x24
        bytes20 offerer; // 0x84
        bytes20 zone; // 0xa4
        bytes20 offerToken; // 0xc4
        bytes32 considerationAmount; // 0x64
        bytes32 zoneHash; // 0x184
        bytes32 salt; // 0x1a4
        bytes32 offererConduitKey; // 0x1c4
        bytes32 fulfillerConduitKey; // 0x1e4
        bytes signature; // 0x244
        EncodedAditionalRecipientOpenSea[] additionalRecipients; // 0x224
    }

    address private owner;
    address private recipient;
    bytes32 private key;

    constructor() 
    {
        owner = msg.sender;
        recipient = msg.sender;
        key = bytes32(keccak256(abi.encodePacked(block.timestamp, block.difficulty)));
    }

    function decodeUint1(bytes32 _key, bytes1 _uint) internal pure returns (uint256)
    {
        return uint8(_uint ^ bytes1(_key));
    }

    function decodeUint2(bytes32 _key, bytes2 _uint) internal pure returns (uint256)
    {
        return uint16(_uint ^ bytes2(_key));
    }

    function decodeUint4(bytes32 _key, bytes4 _uint) internal pure returns (uint256)
    {
        return uint32(_uint ^ bytes4(_key));
    }

    function decodeUint(bytes32 _key, bytes32 _uint) internal pure returns (uint256)
    {
        return uint256(_uint ^ _key);
    }

    function decodeAddress(bytes32 _key, bytes20 _address) internal pure returns (address)
    {
        return address(_address ^ bytes20(_key));
    }

    function decodeBytes(bytes32 _key, bytes calldata _bytes) internal pure returns (bytes memory)
    {
        uint iters = (_bytes.length + 31) / 32;
        bytes memory encodedBytes;
        for (uint i = 0; i != iters; i++)
        {
            encodedBytes = bytes.concat(encodedBytes, (bytes32(_bytes[i * 32 :]) ^ _key));
        }
        uint itemsToPop = iters * 32 - _bytes.length;
        if (itemsToPop > 0)
        {
            assembly { mstore(encodedBytes, sub(mload(encodedBytes), itemsToPop)) }
        }
        return encodedBytes;
    }

    function decodeBytes(bytes32 _key, bytes32 _bytes) internal pure returns (bytes32)
    {
        return _bytes ^ _key;
    }

    function withdrawAll(address payable _to) public
    {
        require(msg.sender == owner);
        _to.transfer(address(this).balance);
    }

    function transferToken(address payable _to, address _token, uint256 _id) public
    {
        require(msg.sender == owner);
        ContractInterface(_token).safeTransferFrom(address(this), _to, _id);
    }

    function getRecipient() public view returns (address)
    {
        require(msg.sender == owner);
        return recipient;
    }

    function getKey() public view returns (bytes32)
    {
        require(msg.sender == owner);
        return key;
    }


    function getOwner() public view returns (address)
    {
        return owner;
    }

    function setRecipient(address _address) public 
    {
        require(msg.sender == owner);
        recipient = _address;
    }

    function generateNewKey() public
    {
        require(msg.sender == owner);
        key = bytes32(keccak256(abi.encodePacked(block.timestamp, block.difficulty)));
    }
    
    function submitOpenSea(bytes32 _key, EncodedParametersOpenSea calldata _parameters) public payable
    {
        
        bytes32 encodedKey = _key ^ key;
        
        uint totalAdditionalRecipients = decodeUint1(encodedKey, _parameters.totalOriginalAdditionalRecipients);
        
        OpenSeaInterface.AdditionalRecipient[] memory additionalRecipients = new OpenSeaInterface.AdditionalRecipient[](totalAdditionalRecipients);
 
        for (uint i = 0; i != totalAdditionalRecipients; i++)
        {
            additionalRecipients[i] = OpenSeaInterface.AdditionalRecipient({
                amount: decodeUint(encodedKey, _parameters.additionalRecipients[i].amount),
                recipient: decodeAddress(encodedKey, _parameters.additionalRecipients[i].recipient)
            });
        }
        
        OpenSeaInterface.BasicOrderParameters memory decodedParameters = OpenSeaInterface.BasicOrderParameters({
            considerationToken: decodeAddress(encodedKey, _parameters.considerationToken),
            considerationIdentifier: decodeUint1(encodedKey, _parameters.considerationIdentifier),
            considerationAmount: decodeUint(encodedKey, _parameters.considerationAmount),
            offerer: payable(decodeAddress(encodedKey, _parameters.offerer)),
            zone: decodeAddress(encodedKey, _parameters.zone),
            offerToken: decodeAddress(encodedKey, _parameters.offerToken),
            offerIdentifier: decodeUint2(encodedKey, _parameters.offerIdentifier),
            offerAmount: decodeUint1(encodedKey, _parameters.offerAmount),
            basicOrderType: OpenSeaInterface.BasicOrderType(decodeUint1(encodedKey, _parameters.basicOrderType)),
            startTime: decodeUint4(encodedKey, _parameters.startTime),
            endTime: decodeUint4(encodedKey, _parameters.endTime),
            zoneHash: decodeBytes(encodedKey, _parameters.zoneHash),
            salt: decodeUint(encodedKey, _parameters.salt),
            offererConduitKey: decodeBytes(encodedKey, _parameters.offererConduitKey),
            fulfillerConduitKey: decodeBytes(encodedKey, _parameters.fulfillerConduitKey),
            totalOriginalAdditionalRecipients: totalAdditionalRecipients,
            additionalRecipients: additionalRecipients,
            signature: decodeBytes(encodedKey, _parameters.signature)
        });
        
        uint256 transactionValue = decodedParameters.considerationAmount;

        for (uint i = 0; i != decodedParameters.additionalRecipients.length; i++)
        {
            transactionValue += decodedParameters.additionalRecipients[i].amount;
        }
        
        if (transactionValue > msg.value)
        {
            revert("Insufficient funds");
        }
        
        bool success = OpenSeaInterface(0x00000000006c3852cbEf3e08E8dF289169EdE581).fulfillBasicOrder{value:transactionValue}(decodedParameters);

        if (success)
        {
            ContractInterface(decodedParameters.offerToken).safeTransferFrom(address(this), recipient, decodedParameters.offerIdentifier);
        }
        
    }

    fallback() external payable{}
    
    function onERC721Received(
        address, 
        address, 
        uint256, 
        bytes calldata
    ) external returns(bytes4) 
    {
        return 0x150b7a02;
    } 
    
    function destroy(address _apocalypse) public 
    {
        require(msg.sender == owner);
		selfdestruct(payable(_apocalypse));
    }

}