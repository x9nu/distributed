package grades

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type studentsHandler struct{}

func RegisterHandlers() {
	handler := new(studentsHandler)
	http.Handle("/students", handler)
	http.Handle("/students/", handler)
}

/*
	/students
	/students/{:id}
	/students/{:id}/grades
*/
func (sh studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	switch len(pathSegments) {
	case 2: // /students
		sh.getAll(w, r)
	case 3: // /students/{:id}
		// id, err := strconv.Atoi(pathSegments[2])
		// if err != nil {
		// 	w.WriteHeader(http.StatusNotFound)
		// 	return
		// }
		id := pathSegments[2]
		sh.getOne(w, r, id)
	case 4: // /students/{:id}/grades
		// id, err := strconv.Atoi(pathSegments[2])
		// if err != nil {
		// 	w.WriteHeader(http.StatusNotFound)
		// 	return
		// }
		id := pathSegments[2]
		sh.addGrade(w, r, id)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (sh studentsHandler) getAll(w http.ResponseWriter, r *http.Request) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	/*
		序列化，两种方法都可以将数据编码为JSON格式：
			- json.Marshal()函数
			- json.NewEncoder()结合Encode()方法
	*/
	data, err := sh.toJSON(students)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) getOne(w http.ResponseWriter, r *http.Request, id string) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := sh.toJSON(student)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed to serialize student: %q", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) addGrade(w http.ResponseWriter, r *http.Request, id string) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}
	var g Grade
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&g)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	student.Grades = append(student.Grades, g)
	w.WriteHeader(http.StatusCreated)
	data, err := sh.toJSON(g)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "applicaiton/json")
	w.Write(data)
}

func (sh studentsHandler) toJSON(obj interface{}) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	err := enc.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize students: %q", err)
	}
	return b.Bytes(), nil
}
