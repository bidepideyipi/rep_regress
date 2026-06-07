extends Control

## 示例游戏2 - 演示与全局GameManager的数据交互
## 展示如何在子游戏中读写全局状态

var game_score: int = 0

@onready var data_display = $CenterContainer/VBox/DataDisplay
@onready var load_button = $CenterContainer/VBox/ButtonContainer/LoadData
@onready var save_button = $CenterContainer/VBox/ButtonContainer/SaveData

## 游戏启动回调
func on_game_start() -> void:
    print("[DemoGame2] 游戏开始")
    game_score = 0

    load_button.pressed.connect(_on_load_data)
    save_button.pressed.connect(_on_save_data)

## 游戏退出回调
func on_game_exit() -> void:
    print("[DemoGame2] 游戏退出")

## 加载全局数据
func _on_load_data() -> void:
    var session_id = GameManager.get_player_data("session_id", "未知")
    var last_score = GameManager.get_player_data("demo_game_2_score", 0)

    print("[DemoGame2] 加载数据 - Session: %s, 上次分数: %d" % [session_id, last_score])

    data_display.text = "会话: %s\n上次分数: %d" % [session_id, last_score]

## 保存数据到全局
func _on_save_data() -> void:
    game_score += randi() % 100
    GameManager.set_player_data("demo_game_2_score", game_score)

    print("[DemoGame2] 保存分数: %d" % game_score)
    data_display.text = "已保存分数: %d" % game_score
