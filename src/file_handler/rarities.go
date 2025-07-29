package file_handler

import (
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"encoding/json"
	"os"
)

func WriteRarities(rarities structs.Rarities, slug string) {
	file, err := json.MarshalIndent(rarities.Tokens, "", " ")
	if err != nil {
		misc.PrintRed("Error writing rarities file: "+err.Error(), true)
		return
	}
	path := constants.RARITIES_DIRECTORY + slug + ".json"
	err = os.WriteFile(path, file, 0644)

	if err != nil {
		misc.PrintRed("Error writing rarities file: "+err.Error(), true)
		return
	}

	misc.PrintGreen("Succesfully saved rarities to file", true)
}
