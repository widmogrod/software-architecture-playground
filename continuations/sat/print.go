package sat

import "fmt"

func (s *solver) Print() {
	fmt.Print(s.printClosures(s.closures))
}

func (s *solver) printClosures(closures Closures) string {
	result := ""
	for _, line := range closures {
		result += fmt.Sprintf("%s \n", s.printPrepositions(line))
	}

	return result
}

func (s *solver) printPrepositions(line []Preposition) string {
	result := ""
	count := len(line)
	for i := 0; i < count; i++ {
		if i > 0 && i < count {
			result += " "
		}

		result += line[i].String()
	}

	return result
}
