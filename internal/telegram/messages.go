package telegram

import (
	"bytes"
	"text/template"

	"github.com/mmcdole/gofeed"
)

type MessageConfig struct {
	FirstName string
}

const welcomeTemplateString = `👋 Hi {{ .FirstName }}!
👨‍⚖️ Ab jetzt versorge ich dich mit den Entscheidungen des Bundesverfassungsgerichts, sobald diese erscheinen 

🔥Und manchmal vielleicht auch vorher... 

Außerdem sage ich dir Bescheid, wenn neue Features zu Verfügen stehen.
Für Feedback gerne an @rd_io wenden 💻.
`

const muellerMsg = `🧑‍⚖️ Es Müllert wieder!
Heute gibt der <b>2. Senat</b> eine Entscheidung in nachstehender Sache bekannt:
<pre>
Die mit einem Antrag auf Erlass einer einstweiligen Anordnung verbundene Verfassungsbeschwerde richtet sich gegen das Urteil des Verfassungsgerichtshofs des Landes Berlin vom 16. November 2022 - VerfGH 154/21 u. a. -, mit dem die Wahlen zum 19. Abgeordnetenhaus von Berlin sowie zu den Bezirksverordnetenversammlungen vom 26. September 2021 im gesamten Wahlgebiet für ungültig erklärt wurden.
</pre>

<i>	2 BvR 2189/22</i>
`

const decisionTemplateString = `🦅 <b>Im Namen des Volkes</b> 🦅
Es wurde nachstehende Entscheidung verkündet:

<i>{{ .Title }}</i>
<pre>{{ .Description }}</pre>

<a href="{{ .Link }}">Zur Entscheidung</a>
`

var (
	welcomeTemplate  *template.Template
	decisionTemplate *template.Template
)

func init() {
	welcomeTemplate, _ = template.New("welcome").Parse(welcomeTemplateString)
	decisionTemplate, _ = template.New("decision").Parse(decisionTemplateString)
}

func getWelcomeMessage(cfg MessageConfig) (string, error) {
	var buf bytes.Buffer
	if err := welcomeTemplate.Execute(&buf, cfg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type decisonCfg struct {
	Title       string
	Description string
	Link        string
}

func buildDecisionMessage(item *gofeed.Item) (string, error) {
	cfg := decisonCfg{
		Title:       item.Title,
		Description: item.Description,
		Link:        item.Link,
	}

	var buf bytes.Buffer
	if err := decisionTemplate.Execute(&buf, cfg); err != nil {
		return "", err
	}

	return buf.String(), nil

}
