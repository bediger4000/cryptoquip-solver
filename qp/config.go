package qp

var scorecard [124]rune

func StringConfiguration(line string) string {
	key := make([]rune, 0, len(line))

	idx := 0

	ln := len(line)
	if ln >= 3 && line[ln-1] == 'l' && line[ln-2] == 'l' && line[ln-3] == '\'' {
		line = line[:ln-3]
		ln = len(line)
	}
	if ln >= 2 && line[ln-1] == 's' && line[ln-2] == '\'' {
		line = line[:ln-2]
		ln = len(line)
	}
	if line[ln-1] == '\'' {
		line = line[:ln-1]
	}

	if len(line) == 0 {
		return ""
	}

	for i, r := range line {
		if i == 0 {
			key = append(key, '0')
			scorecard[idx] = r
			idx++
			continue
		}
		foundit := false
		for j := 0; j < idx; j++ {
			if scorecard[j] == r {
				key = append(key, rune('0'+j))
				foundit = true
				break
			}
		}
		if !foundit {
			scorecard[idx] = r
			key = append(key, rune('0'+idx))
			idx++
		}
	}
	return string(key)
}
