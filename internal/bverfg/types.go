package bverfg

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	// foundingYear holds a two digit year number in which the BVerfG established.
	// used to parse case reference strings for the y2k wrap.
	foundingYear = 51
)

type ProcedureType string

const (
	Art18GG                         = ProcedureType("BvA")
	Parteiverbotsverfahren          = ProcedureType("BvB")
	Wahlpruefungsbeschwerde         = ProcedureType("BvC")
	Praesidentenanklage             = ProcedureType("BvD")
	Organstreit                     = ProcedureType("BvE")
	AbstrakteNormenkontrolle        = ProcedureType("BvF")
	BundLaenderStreit               = ProcedureType("BvG")
	OeffentlichRechtlich            = ProcedureType("BvH")
	Richteranklage                  = ProcedureType("BvJ")
	LandesverfassungsStreitigkeit   = ProcedureType("BvK")
	KonkreteNormenkontrolle         = ProcedureType("BvL")
	Voelkerrechtsbindung            = ProcedureType("BvM")
	Divergenzvorlage                = ProcedureType("BvN")
	VorkonstitutionelleFortgeltung  = ProcedureType("BvO")
	BundesgesetzlichesVerfahren     = ProcedureType("BvP")
	EinstweiligeAnordnung           = ProcedureType("BvQ")
	Verfassungsbeschwerde           = ProcedureType("BvR")
	SonstigesVerfahren              = ProcedureType("BvT")
	Dienstunfaehigkeitsfeststellung = ProcedureType("PBvS")
	Plenarentscheidung              = ProcedureType("PBvU")
	Prozesskostenhilfe              = ProcedureType("PKH")
	Verzoegerungsruege              = ProcedureType("Vz")
)

func (p ProcedureType) RefSign() string {
	return string(p)
}

func (p ProcedureType) String() string {
	switch p {
	case Art18GG:
		return "Verwirkung von Grundrechten"
	case Parteiverbotsverfahren:
		return "Feststellung der Verfassungswidrigkeit einer Partei"
	case Wahlpruefungsbeschwerde:
		return "Beschwerde im Wahlprüfungsverfahren"
	case Praesidentenanklage:
		return "Anklage gegen den Bundespräsidenten"
	case Organstreit:
		return "Verfassungsstreitigkeit zwischen Bundesorganen"
	case AbstrakteNormenkontrolle:
		return "Normenkontrolle auf Antrag von Verfassungsorganen"
	case BundLaenderStreit:
		return "Verfassungsstreitigkeiten zwischen Bund und Ländern"
	case OeffentlichRechtlich:
		return "Öffentlichrechtliche Streitigkeiten"
	case Richteranklage:
		return "Richteranklage"
	case LandesverfassungsStreitigkeit:
		return "Landesverfassungsstreitigkeit kraft landesrechtlicher Zuweisung"
	case KonkreteNormenkontrolle:
		return "Normenkontrolle auf Vorlage von Gerichte"
	case Voelkerrechtsbindung:
		return "Völkerrechtliche Normverifikation"
	case Divergenzvorlage:
		return "Divergenzvorlage"
	case VorkonstitutionelleFortgeltung:
		return "Fortgelten vorkonstitutionellen Rechts als Bundesrecht"
	case BundesgesetzlichesVerfahren:
		return "Sonstiges durch Bundesrecht zugewiesenes Verfahren"
	case EinstweiligeAnordnung:
		return "Verfahren über Anträge im Wege der einstweiligen Anordnung"
	case Verfassungsbeschwerde:
		return "Verfassungsbeschwerde"
	case SonstigesVerfahren:
		return "Sonstiges Verfahren"
	case Dienstunfaehigkeitsfeststellung:
		return "Verfahren über die Beendigung des Amtes eines Richters des BVerfG bei Dienstunfähigkeit oder aus sonstige Gründen"
	case Plenarentscheidung:
		return "Plenarentscheidung"
	case Prozesskostenhilfe:
		return "Prozesskostenhilfe"
	case Verzoegerungsruege:
		return "Verzögerungsrüge"
	}

	return ""
}

type CaseReference struct {
	Senate        uint8
	Type          ProcedureType
	Year          int
	RunningNumber uint
}

var caseRefRegex *regexp.Regexp

func init() {
	caseRefRegex = regexp.MustCompile(`(1|2)\s([A-Za-z]*)\s(\d*)\/(\d{2})`)
}

func (c CaseReference) String() string {
	twoDigitYear := c.Year % 100

	return fmt.Sprintf(
		"%d %v %d/%d",
		c.Senate,
		c.Type.RefSign(),
		c.RunningNumber,
		twoDigitYear,
	)
}

func ParseCaseRef(r string) (CaseReference, error) {
	ref := CaseReference{}

	match := caseRefRegex.FindStringSubmatch(r)
	if len(match) != 5 {
		return ref, fmt.Errorf("invalid case ref match: %v", match)
	}

	senate, err := strconv.Atoi(match[1])
	if err != nil {
		return ref, fmt.Errorf("parse as senate number: %v", match[1])
	}
	ref.Senate = uint8(senate)

	ref.Type = ProcedureType(match[2])

	runningNumber, err := strconv.Atoi(match[3])
	if err != nil {
		return ref, fmt.Errorf("parse as running number: %v", match[3])
	}
	ref.RunningNumber = uint(runningNumber)

	year, err := strconv.Atoi(match[4])
	if err != nil {
		return ref, fmt.Errorf("parse as two digit year: %v", match[4])
	}
	if year >= foundingYear {
		ref.Year = year + 1900
	} else {
		ref.Year = year + 2000
	}

	return ref, nil
}

type Decision struct {
	Ref         CaseReference
	title       string
	link        string
	description string
}
