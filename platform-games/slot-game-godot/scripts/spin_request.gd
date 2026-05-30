class_name SpinRequest
extends RefCounted

var user_id: String
var bet_amount: float
var bet_lines: int
var session_id: String
var is_free_spin: bool = false

func to_dict() -> Dictionary:
	return {
		"user_id": user_id,
		"bet_amount": bet_amount,
		"bet_lines": bet_lines,
		"session_id": session_id,
		"is_free_spin": is_free_spin
	}
