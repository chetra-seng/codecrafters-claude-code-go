package skills

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// ConstructSkillPrompt construct agent prompt telling the skills available based on skill metadata
func ConstructSkillPrompt(skills []skillMeta) string {
	var sb strings.Builder
	sb.WriteString("You have access to the following skills:\n\n")

	for _, skill := range skills {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", skill.Name, skill.Description))
	}

	return sb.String()
}

// LoadSkills return skill metadata based on the given path
func LoadSkills(path string) []skillMeta {

	skills := make([]skillMeta, 0)
	info, err := os.Stat(path)

	if err != nil {
		return skills
	}

	if info.IsDir() {
		md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
		entries, err := os.ReadDir(path)
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				byteCon, err := os.ReadFile(path + "/" + entry.Name())
				if err != nil {
					panic("Skill content can not be read")
				}

				ctx := parser.NewContext()
				var buf bytes.Buffer

				if err := md.Convert(byteCon, &buf, parser.WithContext(ctx)); err != nil {
					log.Fatal(err)
				}

				var meta skillMeta

				d := frontmatter.Get(ctx)

				if d != nil {
					if err := d.Decode(&meta); err != nil {
						log.Fatal(err)
					}
				}
				skills = append(skills, meta)
			}
		}

	}

	return skills
}
