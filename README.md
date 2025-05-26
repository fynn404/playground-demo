# Go Playground

A simple web-based Go playground that allows you to write and execute Go code in your browser.

## Features

- Modern, clean UI with syntax highlighting
- Real-time code execution
- Error handling and output display
- Support for standard Go packages

## Requirements

- Go 1.16 or later

## Running the Playground

1. Clone this repository
2. Navigate to the project directory
3. Run the server:
   ```bash
   go run main.go
   ```
4. Open your browser and visit `http://localhost:8081`

## Usage

1. Write your Go code in the editor
2. Click the "Run Code" button to execute
3. View the output or error messages below the editor

## Security Note

This playground runs code on the server side. In a production environment, you should implement proper security measures such as:
- Code execution timeouts
- Memory limits
- Network access restrictions
- Sandbox environment (e.g., using Docker)

## License

MIT License