package model

import "errors"

var ErrDuplicateLogin = errors.New("duplicate login")
var ErrWrongOrderNumber = errors.New("wrong order number")
var ErrOrderAlreadyExistSameUser = errors.New("order already exist for same user")
var ErrOrderAlreadyExistOtherUser = errors.New("order already exist for other user")
var ErrNoBalance = errors.New("no balance")
