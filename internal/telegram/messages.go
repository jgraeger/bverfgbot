package telegram

import (
	"bytes"
	"text/template"

	"github.com/jgraeger/bverfgbot/internal/bverfg"
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

const firstSenateTodayTpl = `Geheimdienste zittern, Pressekammern schlottern!

Der <b>1. Senat</b>🐻✝️🥦🐺 
gibt heute eine Entscheidung in nachstehender Sache bekannt:
<pre>
{{ .Description }}
</pre>
Aktenzeichen: {{ .RefString }}
`

const secondSenateTodayTpl = `🧑‍⚖️ Es Müllert wieder!
Heute gibt der <b>2. Senat</b> eine Entscheidung in nachstehender Sache bekannt:
<pre>
{{ .Description }}
</pre>

Aktenzeichen: {{ .RefString }}
`

const decisionTemplateString = `🦅 <b>Im Namen des Volkes</b> 🦅
Es wurde nachstehende Entscheidung verkündet:

<i>{{ .Title }}</i>
<pre>{{ .Description }}</pre>

<a href="{{ .Link }}">Zur Entscheidung</a>
`

var (
	welcomeTemplate      *template.Template
	decisionTemplate     *template.Template
	firstSenateTemplate  *template.Template
	secondSenateTemplate *template.Template
)

func init() {
	welcomeTemplate, _ = template.New("welcome").Parse(welcomeTemplateString)
	decisionTemplate, _ = template.New("decision").Parse(decisionTemplateString)
	firstSenateTemplate, _ = template.New("first_senate_daily").Parse(firstSenateTodayTpl)
	secondSenateTemplate, _ = template.New("second_senate_daily").Parse(secondSenateTodayTpl)
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

type upcomingCfg struct {
	Description string
	RefString   string
}

func getUpcomingTemplateFor(senate uint8) *template.Template {
	if senate == 1 {
		return firstSenateTemplate
	}
	return secondSenateTemplate
}

func buildUpcomingDecisionMessage(d bverfg.AnnouncedDecision) (string, error) {
	tpl := getUpcomingTemplateFor(d.Ref.Senate)

	var buf bytes.Buffer
	cfg := upcomingCfg{Description: d.Description, RefString: d.Ref.String()}
	if err := tpl.Execute(&buf, cfg); err != nil {
		return "", err
	}

	return buf.String(), nil
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
