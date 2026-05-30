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
var rtp: float

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
	rtp = response_data.get("rpt", 0.0)
