package util

import (

)

func If(condition bool, trueVal interface{}, falseVal interface{}) interface{}{
	if condition {
		return trueVal
	}else{
		return falseVal
	}
}