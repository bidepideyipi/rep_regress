extends Control

const NUM_REELS = 3
const NUM_ROWS = 3
const SYMBOL_SIZE = 100
const REEL_SPACING = 130
const INITIAL_BALANCE = 1000.0

const PAYLINES = [
	[1, [0, 1, 2]],
	[2, [3, 4, 5]],
	[3, [6, 7, 8]],
	[4, [0, 4, 8]],
	[5, [2, 4, 6]]
]

var http_manager: HttpManager
var symbol_renderer: SymbolRenderer

var current_bet: float = 1.0
var bet_lines: int = 5
var balance: float = INITIAL_BALANCE
var total_win: float = 0.0
var is_spinning: bool = false

var user_id: String = "user_godot_001"
var session_id: String = ""

var symbol_containers: Array = []
var reel_panels: Array = []
var reel_strips: Array = []

var title_label: Label
var balance_label: Label
var bet_label: Label
var win_label: Label
var status_label: Label
var spin_button: Button
var minus_button: Button
var plus_button: Button
var reels_container: Control

var pending_response: SpinResponse = null
var current_win_lines: Array = []
var highlight_nodes: Array = []

func _ready() -> void:
	print("=== Slot Game Starting ===")

	_init_managers()
	_find_nodes()

	call_deferred("_setup_slot_machine_deferred")

func _setup_slot_machine_deferred() -> void:
	_setup_slot_machine()
	_load_initial_symbols()
	_update_ui()
	_connect_signals()
	print("=== Slot Game Ready ===")

func _init_managers() -> void:
	print("Init managers...")
	http_manager = HttpManager.get_instance()
	symbol_renderer = SymbolRenderer.get_instance()

	add_child(http_manager)
	http_manager.spin_completed.connect(_on_spin_completed)
	http_manager.config_completed.connect(_on_config_completed)
	session_id = _generate_session_id()

	http_manager.get_game_config()

func _find_nodes() -> void:
	print("Finding nodes...")

	var vbox = get_node_or_null("VBoxContainer")
	if vbox == null:
		print("ERROR: VBoxContainer not found!")
		return

	title_label = vbox.get_node_or_null("TitleLabel")
	var topbar = vbox.get_node_or_null("TopBar")
	if topbar != null:
		balance_label = topbar.get_node_or_null("BalanceLabel")
		bet_label = topbar.get_node_or_null("BetLabel")
		win_label = topbar.get_node_or_null("WinLabel")

	status_label = vbox.get_node_or_null("StatusLabel")

	var control_panel = vbox.get_node_or_null("ControlPanel")
	if control_panel != null:
		spin_button = control_panel.get_node_or_null("SpinButton")
		minus_button = control_panel.get_node_or_null("MinusButton")
		plus_button = control_panel.get_node_or_null("PlusButton")

	reels_container = vbox.get_node_or_null("ReelsContainer")

func _connect_signals() -> void:
	print("=== Connecting signals ===")
	print("spin_button: ", spin_button)
	print("minus_button: ", minus_button)
	print("plus_button: ", plus_button)
	if spin_button:
		print("Connecting spin_button signal...")
		spin_button.pressed.connect(_on_spin_pressed)
		print("Spin signal connected")
	if minus_button:
		minus_button.pressed.connect(_on_minus_pressed)
	if plus_button:
		plus_button.pressed.connect(_on_plus_pressed)

func _setup_slot_machine() -> void:
	if reels_container == null:
		print("ERROR: ReelsContainer not found!")
		return

	print("Setting up slot machine...")

	var container_width = reels_container.size.x
	if container_width < 400:
		container_width = 450

	var total_width = NUM_REELS * REEL_SPACING + 20
	var start_x = (container_width - total_width) / 2 + 10

	print("Container width: ", container_width, ", Start X: ", start_x)

	for reel in range(NUM_REELS):
		var reel_panel = _create_reel_panel(reel)
		reel_panel.position = Vector2(start_x + reel * REEL_SPACING, 10)
		reels_container.add_child(reel_panel)
		reel_panels.append(reel_panel)

		var clip = reel_panel.get_node("ClipContainer")
		reel_strips.append(clip.get_node("ReelStrip"))

	print("Created ", reel_panels.size(), " reels")

	_create_line_indicators(start_x)

func _create_reel_panel(reel_index: int) -> Control:
	var reel_panel = Control.new()
	reel_panel.name = "Reel_" + str(reel_index)
	reel_panel.custom_minimum_size = Vector2(SYMBOL_SIZE + 20, NUM_ROWS * SYMBOL_SIZE + 20)

	var bg = Panel.new()
	bg.custom_minimum_size = Vector2(SYMBOL_SIZE + 20, NUM_ROWS * SYMBOL_SIZE + 20)
	reel_panel.add_child(bg)

	var clip = Control.new()
	clip.name = "ClipContainer"
	clip.custom_minimum_size = Vector2(SYMBOL_SIZE, NUM_ROWS * SYMBOL_SIZE)
	clip.position = Vector2(10, 10)
	clip.clip_contents = true
	reel_panel.add_child(clip)

	var strip = Control.new()
	strip.name = "ReelStrip"
	strip.custom_minimum_size = Vector2(SYMBOL_SIZE, 0)
	clip.add_child(strip)

	var containers = []
	for row in range(NUM_ROWS):
		var container = Control.new()
		container.custom_minimum_size = Vector2(SYMBOL_SIZE, SYMBOL_SIZE)
		container.position = Vector2(0, row * SYMBOL_SIZE)
		strip.add_child(container)
		containers.append(container)

	symbol_containers.append(containers)

	return reel_panel

func _create_line_indicators(start_x: float) -> void:
	var sidebar = Control.new()
	sidebar.name = "PaylineSidebar"
	sidebar.custom_minimum_size = Vector2(60, NUM_ROWS * SYMBOL_SIZE)
	sidebar.position = Vector2(start_x - 70, 10)
	reels_container.add_child(sidebar)

	var y_offset = 20
	for payline in PAYLINES:
		var line_id = payline[0]
		var positions = payline[1]
		var line_color = _get_line_color(line_id)

		var line_item = _create_payline_item(line_id, positions, line_color)
		line_item.position = Vector2(0, y_offset)
		sidebar.add_child(line_item)

		y_offset += 35

func _create_payline_item(line_id: int, positions: Array, color: Color) -> Control:
	var container = Control.new()
	container.custom_minimum_size = Vector2(50, 30)

	var num_label = Label.new()
	num_label.text = str(line_id)
	num_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	num_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	num_label.add_theme_font_size_override("font_size", 16)
	num_label.add_theme_color_override("font_color", color)
	num_label.custom_minimum_size = Vector2(20, 30)
	num_label.position = Vector2(0, 0)
	container.add_child(num_label)

	var pattern = _create_line_pattern(positions, color)
	pattern.position = Vector2(25, 5)
	container.add_child(pattern)

	return container

func _create_line_pattern(positions: Array, color: Color) -> Control:
	var pattern = Control.new()
	pattern.custom_minimum_size = Vector2(25, 20)

	var dot_size = 6
	var spacing = 8

	for i in range(3):
		var pos = positions[i]
		var row = pos / NUM_REELS

		var dot = ColorRect.new()
		dot.color = color
		dot.custom_minimum_size = Vector2(dot_size, dot_size)
		dot.position = Vector2(i * spacing, row * 6)
		pattern.add_child(dot)

	return pattern

func _load_initial_symbols() -> void:
	var symbol_types = ["cherry", "lemon", "orange", "plum", "grape", "watermelon", "bell", "seven"]

	for reel in range(NUM_REELS):
		for row in range(NUM_ROWS):
			var symbol_type = symbol_types[randi() % symbol_types.size()]
			_display_symbol(reel, row, symbol_type)

	print("Loaded initial symbols")

func _display_symbol(reel: int, row: int, symbol_type: String) -> void:
	if reel < 0 or reel >= symbol_containers.size():
		return
	if row < 0 or row >= symbol_containers[reel].size():
		return

	var container = symbol_containers[reel][row]

	for child in container.get_children():
		child.queue_free()

	var symbol = symbol_renderer.create_symbol(symbol_type, SYMBOL_SIZE)
	if symbol != null:
		container.add_child(symbol)

func _on_spin_pressed() -> void:
	print("=== Spin button pressed ===")
	print("is_spinning: ", is_spinning)
	print("balance: ", balance, ", current_bet: ", current_bet)

	if is_spinning:
		print("Already spinning, ignoring")
		return

	if balance < current_bet:
		_show_error("Insufficient balance!")
		return

	_perform_spin()

func _on_minus_pressed() -> void:
	if not is_spinning and current_bet > 0.5:
		current_bet -= 0.5
		_update_ui()

func _on_plus_pressed() -> void:
	if not is_spinning and current_bet < 100.0:
		current_bet += 0.5
		_update_ui()

func _perform_spin() -> void:
	_clear_highlights()

	is_spinning = true
	pending_response = null
	total_win = 0.0

	if status_label:
		status_label.text = "Spinning..."
	_update_ui()

	var request = SpinRequest.new()
	request.user_id = user_id
	request.bet_amount = current_bet
	request.bet_lines = bet_lines
	request.session_id = session_id
	request.is_free_spin = false

	#print("=== Spin Request ===")
	#print("JSON: ", JSON.stringify(request.to_dict()))

	http_manager.spin(request)

func _on_spin_completed(response: SpinResponse) -> void:
	print("=== Spin Response ===")
	print("JSON: ", JSON.stringify(response.to_dict()))

	if not response.success:
		print("ERROR: ", response.message)
		_show_error(response.message if response.message != "" else "Spin failed")
		_reset_spin_state()
		return

	var reel_result = _transpose_reel_result(response.get_reel_result())
	var win_amount = response.get_win_amount()

	print("Win: ", win_amount)
	print("Result: ", reel_result)

	_start_reel_animation_with_result(reel_result)

	await get_tree().create_timer(3.0).timeout

	_finalize_spin_result(response)

func _on_config_completed(data: Dictionary) -> void:
	print("=== Game Config Response ===")
	print("JSON: ", JSON.stringify(data))

func _transpose_reel_result(row_data: Array) -> Array:
	var transposed = []
	for reel in range(NUM_REELS):
		var reel_symbols = []
		for row in range(NUM_ROWS):
			if row < row_data.size() && reel < row_data[row].size():
				reel_symbols.append(row_data[row][reel])
			else:
				reel_symbols.append("cherry")
		transposed.append(reel_symbols)
	return transposed

func _start_reel_animation_with_result(reel_result: Array) -> void:
	var symbol_types = ["cherry", "lemon", "orange", "plum", "grape", "watermelon", "bell", "seven", "wild", "seven"]

	#print("Starting animation with ", reel_result.size(), " reels")

	for reel in range(NUM_REELS):
		var strip = reel_strips[reel]

		strip.position = Vector2(0, 0)

		for child in strip.get_children():
			strip.remove_child(child)
			child.queue_free()

		symbol_containers[reel].clear()

		var num_random = 15 + reel * 3
		var total_symbols = num_random + NUM_ROWS

		#print("Reel ", reel, ": ", total_symbols, " symbols (", num_random, " random + ", NUM_ROWS, " result)")

		for i in range(total_symbols):
			var container = Control.new()
			container.custom_minimum_size = Vector2(SYMBOL_SIZE, SYMBOL_SIZE)
			container.position = Vector2(0, i * SYMBOL_SIZE)
			strip.add_child(container)
			symbol_containers[reel].append(container)

			var symbol_type: String
			if i < num_random:
				symbol_type = symbol_types[randi() % symbol_types.size()]
			else:
				var row = i - num_random
				if row < reel_result[reel].size():
					symbol_type = reel_result[reel][row]
					#print("  Result at ", i, ": ", symbol_type)
				else:
					symbol_type = symbol_types[randi() % symbol_types.size()]

			var symbol = symbol_renderer.create_symbol(symbol_type, SYMBOL_SIZE)
			if symbol != null:
				container.add_child(symbol)

		var target_y = -(num_random * SYMBOL_SIZE)
		var duration = 1.2 + reel * 0.35

		#print("Reel ", reel, " animating to y=", target_y, " over ", duration, "s")

		var tween = create_tween()
		tween.set_parallel(false)

		tween.tween_property(strip, "position:y", target_y, duration)
		tween.set_ease(Tween.EASE_OUT)
		tween.set_trans(Tween.TRANS_CUBIC)

func _finalize_spin_result(response: SpinResponse) -> void:
	var reel_result = _transpose_reel_result(response.get_reel_result())
	var win_amount = response.get_win_amount()

	for reel in range(min(NUM_REELS, reel_result.size())):
		var strip = reel_strips[reel]

		strip.position = Vector2(0, 0)

		for child in strip.get_children():
			child.queue_free()

		symbol_containers[reel].clear()

		for row in range(NUM_ROWS):
			var container = Control.new()
			container.custom_minimum_size = Vector2(SYMBOL_SIZE, SYMBOL_SIZE)
			container.position = Vector2(0, row * SYMBOL_SIZE)
			strip.add_child(container)
			symbol_containers[reel].append(container)

			if row < reel_result[reel].size():
				var symbol_type = reel_result[reel][row]
				var symbol = symbol_renderer.create_symbol(symbol_type, SYMBOL_SIZE)
				if symbol != null:
					container.add_child(symbol)

	balance -= current_bet
	balance += win_amount
	total_win = win_amount

	is_spinning = false

	_update_ui()

	if status_label:
		status_label.text = "Win: %.2f" % win_amount

	if win_amount > 0:
		current_win_lines = response.get_win_lines()
		_show_win_effect(response)

	print("Done, Balance: ", balance)

func _reset_spin_state() -> void:
	is_spinning = false
	if status_label:
		status_label.text = "Ready"
	_update_ui()

func _show_win_effect(response: SpinResponse) -> void:
	_highlight_winning_symbols(response)

	var win_label = Label.new()
	win_label.text = "WIN!"
	win_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
	win_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
	win_label.add_theme_font_size_override("font_size", 72)
	win_label.add_theme_color_override("font_color", Color.GOLD)
	win_label.z_index = 100

	add_child(win_label)
	win_label.position = Vector2(get_viewport_rect().size.x / 2 - 80, get_viewport_rect().size.y / 2 - 40)

	var tween = create_tween()
	tween.parallel().tween_property(win_label, "modulate:a", 1.0, 0.3).from(0.0)
	tween.parallel().tween_property(win_label, "scale", Vector2(1.5, 1.5), 0.5).from(Vector2(0.5, 0.5))
	tween.tween_interval(1.0)
	tween.tween_property(win_label, "modulate:a", 0.0, 0.3)
	tween.tween_callback(func(): win_label.queue_free())

func _highlight_winning_symbols(response: SpinResponse) -> void:
	var win_lines = response.get_win_lines()
	print("Highlighting ", win_lines.size(), " win lines")

	for win_line in win_lines:
		var win_amount = win_line.get("win_amount", 0.0)
		if win_amount <= 0:
			continue

		var positions = win_line.get("positions", [])
		var line_color = _get_line_color(int(win_line.get("line_id", 0)))
		print("Line ", win_line.get("line_id"), " positions: ", positions, " win: ", win_amount)

		for pos in positions:
			var position_index = int(pos)
			var reel = position_index % NUM_REELS
			var row = position_index / NUM_REELS

			if reel >= 0 && reel < symbol_containers.size() && row >= 0 && row < symbol_containers[reel].size():
				_add_symbol_highlight(reel, row, line_color)

func _add_symbol_highlight(reel: int, row: int, color: Color) -> void:
	if reel >= reel_panels.size():
		return

	var reel_panel = reel_panels[reel]
	var clip = reel_panel.get_node_or_null("ClipContainer")

	if clip == null:
		return

	var highlight = ColorRect.new()
	highlight.color = Color(0, 0, 0, 0)
	highlight.position = Vector2(0, row * SYMBOL_SIZE)
	highlight.custom_minimum_size = Vector2(SYMBOL_SIZE, SYMBOL_SIZE)
	highlight.z_index = 50
	clip.add_child(highlight)
	highlight_nodes.append(highlight)

	var border = Panel.new()
	border.position = Vector2(2, 2)
	border.custom_minimum_size = Vector2(SYMBOL_SIZE - 4, SYMBOL_SIZE - 4)

	var style_box = StyleBoxFlat.new()
	style_box.bg_color = Color(0, 0, 0, 0)
	style_box.border_width_left = 4
	style_box.border_width_top = 4
	style_box.border_width_right = 4
	style_box.border_width_bottom = 4
	style_box.border_color = color

	border.add_theme_stylebox_override("panel", style_box)
	highlight.add_child(border)

	var tween = create_tween()
	tween.set_loops()
	tween.tween_property(border, "modulate:a", 0.5, 0.5)
	tween.tween_property(border, "modulate:a", 1.0, 0.5)

func _get_line_color(line_id: int) -> Color:
	var colors = [
		Color.RED,
		Color.GREEN,
		Color.BLUE,
		Color.YELLOW,
		Color.CYAN,
		Color.MAGENTA,
		Color.ORANGE
	]
	return colors[line_id % colors.size()]

func _clear_highlights() -> void:
	for highlight in highlight_nodes:
		if is_instance_valid(highlight):
			highlight.queue_free()
	highlight_nodes.clear()

func _update_ui() -> void:
	if balance_label:
		balance_label.text = "Credits: %.2f" % balance
	if bet_label:
		bet_label.text = "Bet: %.2f" % current_bet
	if win_label:
		win_label.text = "Win: %.2f" % total_win
	if spin_button:
		spin_button.disabled = is_spinning
	if minus_button:
		minus_button.disabled = is_spinning
	if plus_button:
		plus_button.disabled = is_spinning

func _show_error(message: String) -> void:
	print("ERROR: ", message)
	var error_dialog = AcceptDialog.new()
	error_dialog.title = "Error"
	error_dialog.dialog_text = message

	add_child(error_dialog)
	error_dialog.popup_centered()

	error_dialog.confirmed.connect(func(): error_dialog.queue_free())
	error_dialog.close_requested.connect(func(): error_dialog.queue_free())

func _generate_session_id() -> String:
	return "session_%d_%d" % [Time.get_unix_time_from_system(), randi() % 10000]
