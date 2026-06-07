extends Control

## 大厅主场景 - 管理子游戏列表和加载
## 负责显示可用游戏列表，处理游戏选择和启动

@onready var game_list_container = $VBoxContainer/GameListContainer
@onready var status_label = $VBoxContainer/StatusLabel
@onready var sub_game_viewport = $SubGameViewport
@onready var sub_game_container = $SubGameContainer

var current_loaded_game: Node = null
var game_buttons: Dictionary = {}  # game_id -> Button

func _ready() -> void:
	print("[MainLobby] 大厅初始化")
	_setup_game_list()
	GameLoader.set_viewport(sub_game_viewport)
	GameLoader.game_loaded.connect(_on_game_loaded)
	GameLoader.game_unloaded.connect(_on_game_unloaded)
	GameLoader.load_failed.connect(_on_load_failed)

## 设置游戏列表
func _setup_game_list() -> void:
	var games = SubGameRegistry.get_registered_games()
	print("[MainLobby] 注册的游戏: ", games)

	for game_id in games:
		var game_info = SubGameRegistry.get_game_info(game_id)
		var button = Button.new()
		button.text = game_info.name
		button.custom_minimum_size = Vector2(300, 50)
		button.pressed.connect(_on_game_button_pressed.bind(game_id))
		game_list_container.add_child(button)
		game_buttons[game_id] = button

	status_label.text = "共有 %d 个游戏可用" % games.size()

## 游戏按钮点击处理
func _on_game_button_pressed(game_id: String) -> void:
	print("[MainLobby] 选择游戏: %s" % game_id)

	if current_loaded_game:
		status_label.text = "正在卸载当前游戏..."
		await GameLoader.unload_game()

	status_label.text = "正在加载: %s" % game_id
	GameLoader.load_game(game_id)

## 游戏加载完成回调
func _on_game_loaded(game_node: Node, game_id: String) -> void:
	print("[MainLobby] 游戏加载成功: %s" % game_id)
	current_loaded_game = game_node
	status_label.text = "游戏运行中: %s" % game_id
	sub_game_container.texture = sub_game_viewport.get_texture()

	for gid in game_buttons:
		game_buttons[gid].disabled = (gid == game_id)

## 游戏卸载完成回调
func _on_game_unloaded(game_id: String) -> void:
	print("[MainLobby] 游戏卸载: %s" % game_id)
	current_loaded_game = null
	sub_game_container.texture = null

	for button in game_buttons.values():
		button.disabled = false

## 加载失败回调
func _on_load_failed(error: String) -> void:
	print("[MainLobby] 加载失败: %s" % error)
	status_label.text = "加载失败: %s" % error
	current_loaded_game = null

## 返回大厅快捷键
func _input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_cancel") and current_loaded_game:
		print("[MainLobby] 按下ESC，卸载游戏")
		GameLoader.unload_game()
		get_viewport().set_input_as_handled()
