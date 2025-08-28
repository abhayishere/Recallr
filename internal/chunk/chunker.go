package chunk

import (
	"strings"
	"unicode/utf8"
)

func SplitByParagraph(text string, targetChars, overlapChars int) []string {
	if targetChars <= 0 {
		targetChars = 1200
	}
	if overlapChars <= 0 {
		overlapChars = 200
	}
	paras := splitOnBlankLines(text)

	var chunks []string
	var cur string
	for _, para := range paras {
		if cur == "" {
			cur = para
			continue
		}
		if utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(para) > targetChars {
			chunks = append(chunks, cur)
			if overlapChars > 0 && utf8.RuneCountInString(cur) > overlapChars {
				cur = suffixRunes(cur, overlapChars) + "\n" + para
			}
		} else {
			cur += "\n" + para
		}
	}
	if cur != "" {
		chunks = append(chunks, cur)
	}
	return chunks
}

func splitOnBlankLines(text string) []string {
	raw := strings.Split(text, "\n")
	var out []string
	var buf []string
	for _, line := range raw {
		if strings.TrimSpace(line) == "" {
			if len(buf) > 0 {
				out = append(out, strings.Join(buf, "\n"))
				buf = nil
			}
		} else {
			buf = append(buf, line)
		}
	}
	if len(buf) > 0 {
		out = append(out, strings.Join(buf, "\n"))
	}
	if len(out) == 0 && strings.TrimSpace(text) != "" {
		return []string{strings.TrimSpace(text)}
	}
	return out
}

func suffixRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[len(r)-n:])
}
