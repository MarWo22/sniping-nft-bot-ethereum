package rarities

import (
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"sort"
	"strconv"
)

type sortStruct struct {
	Idx    int
	Rarity float64
}

func GetRarities(tokens map[int]structs.Token, supply int) structs.Rarities {
	addTraitCount(&tokens)
	traits := countTraits(tokens)
	validateCollection(traits)
	addEmptyTraits(&traits, supply)
	sortStructs := calculateRarityScores(&tokens, traits, supply)
	ranks := calculateRanks(sortStructs, &tokens)

	return structs.Rarities{Ranks: ranks, Tokens: tokens}
}

func addTraitCount(tokens *map[int]structs.Token) {
	for key, token := range *tokens {
		token.Traits = append(token.Traits, structs.Attributes{
			TraitType: "traitCount",
			Value:     strconv.Itoa(len(token.Traits)),
		})
		(*tokens)[key] = token
	}
}

func countTraits(tokens map[int]structs.Token) map[string]map[interface{}]int {
	traits := make(map[string]map[interface{}]int)

	for _, token := range tokens {
		for _, trait := range token.Traits {
			if _, ok := traits[trait.TraitType]; !ok {
				traits[trait.TraitType] = make(map[interface{}]int)
			}
			traits[trait.TraitType][trait.Value]++
		}
	}
	return traits
}

func addEmptyTraits(traits *map[string]map[interface{}]int, supply int) {
	for traitTypeName, traitType := range *traits {
		sum := 0
		for _, trait := range traitType {
			sum += trait
		}
		(*traits)[traitTypeName]["MISSING_TRAIT"] = supply - sum
	}
}

func calculateRarityScores(tokens *map[int]structs.Token, traits map[string]map[interface{}]int, supply int) []sortStruct {

	traitCategories := len(traits)
	sortStructs := []sortStruct{}
	for id, token := range *tokens {
		rarityScore := 0.0
		for key, traitType := range traits {
			trait := containsTrait(token.Traits, key)

			if trait == "" {
				rarityScore += (float64(supply) / float64(traitType["MISSING_TRAIT"])) / ((float64(supply) / 1000000) * float64(traitCategories) * float64(len(traitType)))
			} else {
				rarityScore += (float64(supply) / float64(traitType[trait])) / ((float64(supply) / 1000000) * float64(traitCategories) * float64(len(traitType)))
			}
		}
		token.Rarity = rarityScore
		(*tokens)[id] = token
		sortStructs = append(sortStructs, sortStruct{id, rarityScore})
	}
	return sortStructs
}

func containsTrait(traits []structs.Attributes, trait string) interface{} {
	for _, traitVal := range traits {
		if traitVal.TraitType == trait {
			return traitVal.Value
		}
	}
	return ""
}

func calculateRanks(sortStructs []sortStruct, tokens *map[int]structs.Token) []int {
	sort.SliceStable(sortStructs, func(i, j int) bool {
		return sortStructs[i].Rarity > sortStructs[j].Rarity
	})
	ranks := make([]int, len(sortStructs))

	previousRank := 0
	for idx, sortStruct := range sortStructs {
		ranks[idx] = sortStruct.Idx
		token := (*tokens)[sortStruct.Idx]
		if previousRank != 0 && sortStruct.Rarity == sortStructs[idx-1].Rarity {
			token.Rank = previousRank
		} else {
			token.Rank = idx + 1
			previousRank = idx + 1
		}
		(*tokens)[sortStruct.Idx] = token
	}

	return ranks
}

// Makes sure we don't continue on an unrevealed collection
func validateCollection(traits map[string]map[interface{}]int) {
	if len(traits) <= 2 {

		for _, element := range traits {
			if len(element) != 1 {
				return
			}
		}

		misc.PrintRed("Collection has only one trait, aborting", true)
		misc.ExitProgram()
	}
}
