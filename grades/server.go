package grades

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type studentHandler struct{}

func RegisterHandlers() {
	handler := new(studentHandler)
	http.Handle("/students", handler)
	http.Handle("/student/", handler)
}

/*
	/students
	/student/{id}
	/student/{id}/{grades}
*/
func (sh studentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) == 2 && pathSegments[1] == "students" {
		sh.getAll(w, r)
	} else if len(pathSegments) == 3 && pathSegments[1] == "student" {
		id := pathSegments[2]
		sh.getOne(w, r, id)
	} else if len(pathSegments) == 3 && pathSegments[3] == "grades" {
		id := pathSegments[2]
		sh.addGrade(w, r, id)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (sh studentHandler) getAll(w http.ResponseWriter, r *http.Request) {
	studentMutex.Lock()
	defer studentMutex.Unlock()

	/*
		序列化，两种方法都可以将数据编码为JSON格式：
			- json.Marshal()函数
			- json.NewEncoder()结合Encode()方法
	*/
	data, err := json.Marshal(&students)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentHandler) getOne(w http.ResponseWriter, r *http.Request, id string) {
	studentMutex.Lock()
	defer studentMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := json.Marshal(student)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentHandler) addGrade(w http.ResponseWriter, r *http.Request, id string) {
	studentMutex.Lock()
	defer studentMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := json.Marshal(student.Grades)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
