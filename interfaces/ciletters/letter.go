//go:build !solution

package ciletters

import (
	"bytes"
	"strings"
	"text/template"
)

const tmplText = `Your pipeline #{{.Pipeline.ID}}{{if eq .Pipeline.Status "failed"}} has failed!{{else}} passed!{{end}}
    Project:      {{.Project.GroupID}}/{{.Project.ID}}
    Branch:       🌿 {{.Branch}}
    Commit:       {{shortHash .Commit.Hash}} {{.Commit.Message}}
    CommitAuthor: {{.Commit.Author}}{{if .Pipeline.FailedJobs}}{{ range .Pipeline.FailedJobs}}
        Stage: {{ .Stage}}, Job {{ .Name}}
{{lastLines .RunnerLog}}
{{ end }}
{{ else }}
{{ end }}`

func shortHash(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func lastLines(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 10 {
		lines = lines[len(lines)-10:]
	}

	for i := range lines {
		lines[i] = strings.Repeat(" ", 12) + lines[i]
	}

	return strings.Join(lines, "\n")
}

func MakeLetter(n *Notification) (string, error) {
	funcMap := template.FuncMap{
		"shortHash": shortHash,
		"lastLines": lastLines,
	}

	tmpl, err := template.New("letter").Funcs(funcMap).Parse(tmplText)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, n); err != nil {
		return "", err
	}

	return buf.String()[:len(buf.String())-1], nil
}
