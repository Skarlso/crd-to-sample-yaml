package sanitize

import (
	"bytes"
)

func Sanitize(content []byte) ([]byte, error) {
	// bail early if there are no template characters in the CRD
	if !bytes.Contains(content, []byte("{{")) {
		return content, nil
	}

	var result [][]byte //nolint:prealloc // no idea what the size will be

	for _, line := range bytes.Split(content, []byte("\n")) {
		if bytes.HasPrefix(bytes.TrimLeft(line, " "), []byte("{{")) {
			// skip lines that begin with {{
			continue
		}

		// replace {{ }} mid-lines with dummy value
		if begin := bytes.Index(line, []byte("{{")); begin != -1 {
			end := bytes.Index(line, []byte("}}"))
			if end == -1 {
				// we don't have a closing bracket so apply the line.
				result = append(result, line)

				continue
			}

			var newLine []byte
			newLine = append(newLine, line[:begin]...)
			newLine = append(newLine, []byte("replaced")...)
			newLine = append(newLine, line[end+2:]...)
			line = newLine
		}

		result = append(result, line)
	}

	return bytes.Join(result, []byte("\n")), nil
}
