extends Node

## 子游戏注册表 - 管理所有可加载的子游戏信息
## 游戏需要在这里注册才能被大厅识别和加载

## 游戏信息结构
## {
##   "game_id": {
##     "name": "游戏显示名称",
##     "scene_path": "场景文件路径",
##     "description": "游戏描述",
##     "author": "作者"
##   }
## }

var _registered_games: Dictionary = {}

func _ready() -> void:
	_register_default_games()
	print("[SubGameRegistry] 注册表初始化完成")

## 注册默认游戏
func _register_default_games() -> void:
	register_game("demo_game_1", {
		name = "示例游戏1",
		scene_path = "res://scenes/sub_games/demo_game_1.tscn",
		description = "一个简单的示例游戏",
		author = "System"
	})

	register_game("demo_game_2", {
		name = "示例游戏2",
		scene_path = "res://scenes/sub_games/demo_game_2.tscn",
		description = "另一个示例游戏",
		author = "System"
	})

## 注册新游戏
func register_game(game_id: String, info: Dictionary) -> void:
	if _registered_games.has(game_id):
		print("[SubGameRegistry] 游戏已存在，覆盖: %s" % game_id)

	_registered_games[game_id] = {
		"name": info.get("name", game_id),
		"scene_path": info.get("scene_path", ""),
		"description": info.get("description", ""),
		"author": info.get("author", "Unknown")
	}
	print("[SubGameRegistry] 注册游戏: %s -> %s" % [game_id, _registered_games[game_id].name])

## 获取所有已注册游戏的ID列表
func get_registered_games() -> Array:
	return _registered_games.keys()

## 获取游戏信息
func get_game_info(game_id: String) -> Dictionary:
	return _registered_games.get(game_id, {})

## 检查游戏是否已注册
func is_registered(game_id: String) -> bool:
	return _registered_games.has(game_id)

## 注销游戏
func unregister_game(game_id: String) -> void:
	if _registered_games.erase(game_id):
		print("[SubGameRegistry] 注销游戏: %s" % game_id)
	else:
		print("[SubGameRegistry] 游戏不存在: %s" % game_id)
