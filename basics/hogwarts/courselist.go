//go:build !solution

package hogwarts

func buildGraph(prereqs map[string][]string) (map[string][]string, []string) {
	reversed := make(map[string][]string)
	courseSet := make(map[string]bool)

	for course, pres := range prereqs {
		courseSet[course] = true
		for _, pre := range pres {
			reversed[pre] = append(reversed[pre], course)
			courseSet[pre] = true
		}
	}

	courses := make([]string, 0, len(courseSet))
	for course := range courseSet {
		courses = append(courses, course)
	}

	return reversed, courses
}

func dfs(course string, reversed map[string][]string, visited map[string]int, result *[]string) {
	if visited[course] == 1 {
		panic("cycle detected")
	}

	if visited[course] == 2 {
		return
	}

	visited[course] = 1

	for _, dependent := range reversed[course] {
		dfs(dependent, reversed, visited, result)
	}

	visited[course] = 2

	*result = append([]string{course}, *result...)
}

func GetCourseList(prereqs map[string][]string) []string {
	reversed, courses := buildGraph(prereqs)
	visited := make(map[string]int)
	result := []string{}

	for _, course := range courses {
		if visited[course] == 0 {
			dfs(course, reversed, visited, &result)
		}
	}

	return result
}
