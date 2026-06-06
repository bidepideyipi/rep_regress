class_name SpinResponse
extends RefCounted

var success: bool
var message: String = ""
var data: Dictionary = {}

func to_dict() -> Dictionary:
	return {
		"success": success,
		"message": message,
		"data": data
	}

func _from_json(json: Dictionary) -> void:
	success = json.get("success", false)
	message = json.get("message", "")
	data = json.get("data", {})

func get_session_id() -> String:
	return data.get("session_id", "")

func get_user_id() -> String:
	return data.get("user_id", "")

func get_game_id() -> String:
	return data.get("game_id", "")

func get_bet_amount() -> float:
	return data.get("bet_amount", 0.0)

func get_bet_lines() -> int:
	return data.get("bet_lines", 0)

func get_bet_per_line() -> float:
	return data.get("bet_per_line", 0.0)

func get_win_amount() -> float:
	return data.get("win_amount", 0.0)

func get_net_result() -> float:
	return data.get("net_result", 0.0)

func get_is_free_spin() -> bool:
	return data.get("is_free_spin", false)

func get_reel_result() -> Array:
	return data.get("reel_result", [])

func get_win_lines() -> Array:
	return data.get("win_lines", [])

func get_bonus_feature() -> String:
	return data.get("bonus_feature", "")

func get_processing_time_ms() -> float:
	return data.get("processing_time_ms", 0.0)

func get_timestamp() -> String:
	return data.get("timestamp", "")

func get_balance() -> float:
	return data.get("balance", 0.0)
