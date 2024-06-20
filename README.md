# dedao-ebook-srv

This project provides an Atom RSS feed for new eBooks from Dedao.

## Getting Started

### Prerequisites

- Go programming language installed on your machine.

### Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/d100972/dedao-ebook-srv.git
    cd dedao-ebook-srv
    ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

### Running the Application

To run the application, execute the following command:
```bash
go run main.go
```

### Local Testing

You can test the application locally by accessing the following URL in your browser:
```
http://127.0.0.1:8080/feeds/dedao.atom
```

### Building the Application

To build the application, you can use the provided `start.sh` script. This script will detect your operating system and build the application accordingly.

```bash
./start.sh
```

The application will run in the background, and logs will be written to `app.log`.

## Contributing

If you would like to contribute to this project, please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
```