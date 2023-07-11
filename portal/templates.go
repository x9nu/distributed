package portal

import "text/template"

var rootTemplate *template.Template

func ImportTemplates() error {
	var err error
	// 在cmd/portal/main.go使用该函数，所以是它的相对路径
	rootTemplate, err = template.ParseFiles(
		"../../portal/student.html",
		"../../portal/students.html",
	)
	if err != nil {
		return err
	}
	return nil
}
