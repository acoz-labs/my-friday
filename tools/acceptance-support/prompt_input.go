package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

const builderDescriptionPrefix = "Help define, scaffold, inspect, validate, and test"

var skillCatalogRecordPattern = regexp.MustCompile(`^[ \t]*-[ \t]+([^:[:space:]]+):[ \t]+(.+)[ \t]+\(file:[ \t]+([^()[:space:]]+)\)[ \t]*$`)
var skillRootBindingPattern = regexp.MustCompile("^[ \\t]*-[ \\t]+`(r[0-9]+)`[ \\t]+=[ \\t]+`([^`]+)`[ \\t]*$")

type promptInputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type promptInputItem struct {
	Role    string               `json:"role"`
	Content []promptInputContent `json:"content"`
}

func validateBuilderPromptInput(input io.Reader, skillRoot, skill, promptSHA256 string) error {
	if !filepath.IsAbs(skillRoot) || filepath.Clean(skillRoot) != skillRoot || filepath.Base(skillRoot) != "skills" {
		return errors.New("invalid builder skill root")
	}
	canonicalRoot, err := filepath.EvalSymlinks(skillRoot)
	if err != nil {
		return errors.New("invalid builder skill root")
	}
	if skill != "capability-builder" {
		return errors.New("invalid builder skill name")
	}
	if len(promptSHA256) != 64 {
		return errors.New("invalid builder prompt digest")
	}
	if _, err := hex.DecodeString(promptSHA256); err != nil {
		return errors.New("invalid builder prompt digest")
	}
	body, err := io.ReadAll(io.LimitReader(input, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return errors.New("invalid builder prompt input")
	}
	var items []promptInputItem
	if json.Unmarshal(body, &items) != nil || len(items) == 0 {
		return errors.New("invalid builder prompt input")
	}
	var developerLines []string
	var prompt string
	for _, item := range items {
		for _, content := range item.Content {
			if content.Type != "input_text" {
				continue
			}
			if item.Role == "developer" {
				developerLines = append(developerLines, strings.Split(content.Text, "\n")...)
			}
			if item.Role == "user" {
				prompt = content.Text
			}
		}
	}
	wantPrefix := "$" + skill
	if prompt != wantPrefix && !strings.HasPrefix(prompt, wantPrefix+" ") {
		return errors.New("builder prompt is not a literal skill invocation")
	}
	sum := sha256.Sum256([]byte(prompt))
	if hex.EncodeToString(sum[:]) != promptSHA256 {
		return errors.New("builder prompt digest mismatch")
	}
	return validateBuilderCatalogRecord(developerLines, canonicalRoot, skill)
}

func validateBuilderCatalogRecord(lines []string, canonicalRoot, skill string) error {
	var records [][]string
	for _, line := range lines {
		match := skillCatalogRecordPattern.FindStringSubmatch(line)
		if len(match) == 4 && match[1] == skill {
			records = append(records, match)
		}
	}
	if len(records) != 1 {
		return errors.New("builder skill catalog record is missing or duplicated")
	}
	record := records[0]
	if !strings.HasPrefix(record[2], builderDescriptionPrefix) {
		return errors.New("builder skill description is not model-visible")
	}
	directPath := filepath.Join(canonicalRoot, skill, "SKILL.md")
	if record[3] == directPath {
		return nil
	}
	aliasPathPattern := regexp.MustCompile(`^(r[0-9]+)/` + regexp.QuoteMeta(skill) + `/SKILL\.md$`)
	aliasPath := aliasPathPattern.FindStringSubmatch(record[3])
	if len(aliasPath) != 2 {
		return errors.New("builder skill file is not model-visible")
	}
	var bindings []string
	for _, line := range lines {
		match := skillRootBindingPattern.FindStringSubmatch(line)
		if len(match) == 3 && match[1] == aliasPath[1] {
			bindings = append(bindings, match[2])
		}
	}
	if len(bindings) != 1 || bindings[0] != canonicalRoot {
		return errors.New("builder skill root alias is missing, duplicated, or conflicting")
	}
	return nil
}
