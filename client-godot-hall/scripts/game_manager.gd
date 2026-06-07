extends Node

## 游戏管理器 - 全局游戏状态管理
## 处理游戏间的数据传递和全局状态

## 全局游戏数据
var player_data: Dictionary = {}
var shared_state: Dictionary = {}
var session_id: String = ""

## 游戏切换回调
signal before_game_switch(from_game: String, to_game: String)
signal after_game_switch(game_id: String)

func _ready() -> void:
    _init_session()
    print("[GameManager] 游戏管理器初始化")

## 初始化会话
func _init_session() -> void:
    session_id = Time.get_datetime_string_from_system().replace("-", "").replace(":", "").replace(" ", "_")
    print("[GameManager] 会话ID: %s" % session_id)

## 设置玩家数据
func set_player_data(key: String, value: Variant) -> void:
    player_data[key] = value

## 获取玩家数据
func get_player_data(key: String, default_value: Variant = null) -> Variant:
    return player_data.get(key, default_value)

## 设置共享状态
func set_shared_state(key: String, value: Variant) -> void:
    shared_state[key] = value

## 获取共享状态
func get_shared_state(key: String, default_value: Variant = null) -> Variant:
    return shared_state.get(key, default_value)

## 通知游戏切换
func notify_game_switch(from_game: String, to_game: String) -> void:
    print("[GameManager] 游戏切换: %s -> %s" % [from_game, to_game])
    before_game_switch.emit(from_game, to_game)
    await get_tree().process_frame
    after_game_switch.emit(to_game)
