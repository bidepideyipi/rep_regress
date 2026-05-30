extends Node

class_name HttpManager

static var _instance: HttpManager = null
var base_url: String = "http://127.0.0.1:8081"
var http_request: HTTPRequest

signal spin_completed(response: SpinResponse)
signal config_completed(data: Dictionary)
signal health_completed(data: Dictionary)

static func get_instance() -> HttpManager:
	if _instance == null:
		_instance = HttpManager.new()
	return _instance

func _ready() -> void:
	process_mode = Node.PROCESS_MODE_ALWAYS

	if http_request == null:
		http_request = HTTPRequest.new()
		add_child(http_request)
		http_request.request_completed.connect(_on_request_completed)

func set_base_url(url: String) -> void:
	base_url = url

func get_base_url() -> String:
	return base_url

var _current_request_type: String = ""
var _request_data: Dictionary = {}

func spin(request: SpinRequest) -> void:
	if http_request == null:
		_ready()

	var url = base_url + "/api/game/spin"
	var json_string = JSON.stringify(request.to_dict())
	var headers = ["Content-Type: application/json"]

	_current_request_type = "spin"
	_request_data = {}

	http_request.request(url, headers, HTTPClient.METHOD_POST, json_string)

func get_game_config() -> void:
	if http_request == null:
		_ready()

	var url = base_url + "/api/game/config"
	var headers = ["Content-Type: application/json"]

	_current_request_type = "config"
	_request_data = {}

	http_request.request(url, headers, HTTPClient.METHOD_GET)

func health_check() -> void:
	if http_request == null:
		_ready()

	var url = base_url + "/health"
	var headers = ["Content-Type: application/json"]

	_current_request_type = "health"
	_request_data = {}

	http_request.request(url, headers, HTTPClient.METHOD_GET)

func _on_request_completed(result: int, _response_code: int, _headers: PackedStringArray, body: PackedByteArray) -> void:
	if result != HTTPRequest.RESULT_SUCCESS:
		print("HTTP Request failed with result: ", result)
		_handle_error("Request failed")
		return

	var json_string = body.get_string_from_utf8()
	var json = JSON.new()
	var parse_result = json.parse(json_string)

	if parse_result != OK:
		print("Failed to parse JSON response")
		_handle_error("Invalid JSON response")
		return

	var response_data = json.data

	match _current_request_type:
		"spin":
			var response = SpinResponse.new()
			response._from_json(response_data)
			spin_completed.emit(response)
		"config":
			config_completed.emit(response_data)
		"health":
			health_completed.emit(response_data)
		_:
			print("Unknown request type: ", _current_request_type)

func _handle_error(message: String) -> void:
	match _current_request_type:
		"spin":
			var response = SpinResponse.new()
			response.success = false
			response.message = message
			spin_completed.emit(response)
		"config":
			config_completed.emit({"success": false, "message": message})
		"health":
			health_completed.emit({"success": false, "message": message})
