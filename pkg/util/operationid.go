package util

import (
	"strings"

	"github.com/pb33f/libopenapi/orderedmap"
)

func ToOperationId(method string, url string) string {
	operationID := strings.Replace(url, "/api/", "", 1)
	operationID = strings.Replace(operationID, "/rest/", "", 1)
	operationID = strings.Replace(operationID, "/oauth2/", "/OAuth2/", 1)
	operationID = convertPathParameterToSingularIfFollowedByVariable(operationID)
	//operationID = URLRemovePathParams(operationID)
	operationID = URLPathParamAddByPrefix(operationID)

	// get version and remove it from the operationID
	version := ParseURLAPIVersion(url)
	operationID = strings.Replace(operationID, "v"+version+"/", "", 1)
	operationID = strings.Replace(operationID, "api/"+version+"/", "", 1)
	operationID = strings.Replace(operationID, "*", "", 1)
	operationID = strings.Replace(operationID, ".", "", -1)

	return strings.ToLower(method) + CapitalizeAfterChars(operationID, []int32{'/', '-', ':'}, true) + "V" + version
}

func convertPathParameterToSingularIfFollowedByVariable(path string) string {
	sections := strings.Split(path, "/")
	for i := 0; i < len(sections)-1; i++ {
		nextSection := sections[i+1]
		currentSection := sections[i]

		if strings.HasPrefix(nextSection, "{") {
			currentSection = toSingular(currentSection)
		}

		sections[i] = currentSection
	}
	return strings.Join(sections, "/")
}

var irregularForms = map[string]string{
	// no plural form
	`adulthood`:      `adulthood`,
	`advice`:         `advice`,
	`agenda`:         `agenda`,
	`aid`:            `aid`,
	`aircraft`:       `aircraft`,
	`alcohol`:        `alcohol`,
	`ammo`:           `ammo`,
	`analytics`:      `analytics`,
	`anime`:          `anime`,
	`athletics`:      `athletics`,
	`audio`:          `audio`,
	`bison`:          `bison`,
	`blood`:          `blood`,
	`bream`:          `bream`,
	`buffalo`:        `buffalo`,
	`butter`:         `butter`,
	`carp`:           `carp`,
	`cash`:           `cash`,
	`chassis`:        `chassis`,
	`chess`:          `chess`,
	`clothing`:       `clothing`,
	`cod`:            `cod`,
	`commerce`:       `commerce`,
	`cooperation`:    `cooperation`,
	`corps`:          `corps`,
	`debris`:         `debris`,
	`diabetes`:       `diabetes`,
	`digestion`:      `digestion`,
	`elk`:            `elk`,
	`energy`:         `energy`,
	`equipment`:      `equipment`,
	`excretion`:      `excretion`,
	`expertise`:      `expertise`,
	`firmware`:       `firmware`,
	`flounder`:       `flounder`,
	`fun`:            `fun`,
	`gallows`:        `gallows`,
	`garbage`:        `garbage`,
	`graffiti`:       `graffiti`,
	`hardware`:       `hardware`,
	`headquarters`:   `headquarters`,
	`health`:         `health`,
	`herpes`:         `herpes`,
	`highjinks`:      `highjinks`,
	`homework`:       `homework`,
	`housework`:      `housework`,
	`information`:    `information`,
	`jeans`:          `jeans`,
	`justice`:        `justice`,
	`kudos`:          `kudos`,
	`labour`:         `labour`,
	`literature`:     `literature`,
	`machinery`:      `machinery`,
	`mackerel`:       `mackerel`,
	`mail`:           `mail`,
	`media`:          `media`,
	`mews`:           `mews`,
	`moose`:          `moose`,
	`music`:          `music`,
	`mud`:            `mud`,
	`manga`:          `manga`,
	`news`:           `news`,
	`only`:           `only`,
	`personnel`:      `personnel`,
	`pike`:           `pike`,
	`plankton`:       `plankton`,
	`pliers`:         `pliers`,
	`police`:         `police`,
	`pollution`:      `pollution`,
	`premises`:       `premises`,
	`rain`:           `rain`,
	`research`:       `research`,
	`rice`:           `rice`,
	`salmon`:         `salmon`,
	`scissors`:       `scissors`,
	`series`:         `series`,
	`sewage`:         `sewage`,
	`shambles`:       `shambles`,
	`shrimp`:         `shrimp`,
	`software`:       `software`,
	`staff`:          `staff`,
	`swine`:          `swine`,
	`tennis`:         `tennis`,
	`traffic`:        `traffic`,
	`transportation`: `transportation`,
	`trout`:          `trout`,
	`tuna`:           `tuna`,
	`wealth`:         `wealth`,
	`welfare`:        `welfare`,
	`whiting`:        `whiting`,
	`wildebeest`:     `wildebeest`,
	`wildlife`:       `wildlife`,
	`you`:            `you`,
}

var pluralSuffixes = orderedmap.New[string, string]()

func toSingular(word string) string {
	if pluralSuffixes.Len() == 0 {
		pluralSuffixes.Set("ies", "y")   // e.g., "cities" -> "city"
		pluralSuffixes.Set("ves", "f")   // e.g., "wolves" -> "wolf"
		pluralSuffixes.Set("oes", "o")   // e.g., "heroes" -> "hero"
		pluralSuffixes.Set("ses", "s")   // e.g., "masses" -> "mass"
		pluralSuffixes.Set("xes", "x")   // e.g., "foxes" -> "fox"
		pluralSuffixes.Set("ches", "ch") // e.g., "batches" -> "batch"
		pluralSuffixes.Set("shes", "sh") // e.g., "wishes" -> "wish"
		pluralSuffixes.Set("s", "")      // e.g., "cats" -> "cat"
	}

	// irregular forms
	if val, ok := irregularForms[word]; ok {
		return val
	}

	// regular forms
	for item := pluralSuffixes.Oldest(); item != nil; item = item.Next() {
		if strings.HasSuffix(word, item.Key) {
			return strings.TrimSuffix(word, item.Key) + item.Value
		}
	}

	return word
}
