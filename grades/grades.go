package grades

import (
	"fmt"
	"sync"
)

// Student
type Student struct {
	ID        string
	FirstName string
	LastName  string
	Grades    []Grade
}

func (s Student) Average() float32 {
	var result float32
	for _, grade := range s.Grades {
		result += grade.Score
	}
	return result / float32(len((s.Grades)))
}

type Students []Student

var (
	students     Students
	studentMutex *sync.Mutex
)

func (ss Students) GetByID(id string) (*Student, error) {
	for i, _ := range ss {
		if ss[i].ID == id {
			return &ss[i], nil
		}
	}
	return nil, fmt.Errorf("student with ID %d not found", id)
}

// Grade
type GradeType string

const (
	GradeQuiz = GradeType("Quiz")
	GradeTest = GradeType("Test")
	GradeExam = GradeType("Exam")
)

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}
