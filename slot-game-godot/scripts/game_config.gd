class_name GameConfig
extends RefCounted

var game_id: String
var game_name: String
var version: String
var symbols_count: int
var reels_count: int
var pay_lines: int
var min_bet: float
var max_bet: float
var symbol_paytable: Array = []

func _from_json(json: Dictionary) -> void:
	var response_data = json.get("data", {})
	game_id = response_data.get("game_id", "")
	game_name = response_data.get("game_name", "")
	version = response_data.get("version", "")
	symbols_count = response_data.get("symbols_count", 0)
	reels_count = response_data.get("reels_count", 0)
	pay_lines = response_data.get("pay_lines", 0)
	min_bet = response_data.get("min_bet", 0.0)
	max_bet = response_data.get("max_bet", 0.0)
	symbol_paytable = response_data.get("symbol_paytable", [])

func get_symbol_multiplier(symbol_id: String, match_count: int) -> float:
	for symbol_data in symbol_paytable:
		if symbol_data.get("symbol_id") == symbol_id:
			var multipliers = symbol_data.get("multipliers", [])
			for mult in multipliers:
				if mult.get("match_count") == match_count:
					return mult.get("multiplier", 0.0)
	return 0.0

func get_symbol_name(symbol_id: String) -> String:
	for symbol_data in symbol_paytable:
		if symbol_data.get("symbol_id") == symbol_id:
			return symbol_data.get("symbol_name", symbol_id)
	return symbol_id

func get_symbol_type(symbol_id: String) -> String:
	for symbol_data in symbol_paytable:
		if symbol_data.get("symbol_id") == symbol_id:
			return symbol_data.get("symbol_type", "normal")
	return "normal"
