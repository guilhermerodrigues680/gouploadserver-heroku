package handler

import "errors"

var (
	ErrFileIsNotRegular = errors.New("File is not regular")
	ErrFileIsNotDir     = errors.New("File is not dir")
	ErrCreateTemplate   = errors.New("Create template error")
	ErrExecuteTemplate  = errors.New("Execute template error")
)
