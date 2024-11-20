<h1 align="center">go-dork-google</h1>

<p align="center">
   ___              ___           _           ___                  _
  / _ \___         /   \___  _ __| | __      / _ \___   ___   __ _| | ___
 / /_\/ _ \ _____ / /\ / _ \| '__| |/ /____ / /_\/ _ \ / _ \ / _` | |/ _ \
/ /_\\ (_) |_____/ /_// (_) | |  |   <_____/ /_\\ (_) | (_) | (_| | |  __/
\____/\___/     /___,' \___/|_|  |_|\_\    \____/\___/ \___/ \__, |_|\___|
                                                             |___/        </p>

<p align="center">
  <a href="https://golang.org">
    <img src="https://img.shields.io/badge/Made%20with-Go-1f425f.svg" alt="Made with Go">
  </a>
  <a href="https://github.com/Abhinandan-Khurana/go-dork-google/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/Abhinandan-Khurana/go-dork-google" alt="License">
  </a>
  <a href="https://github.com/Abhinandan-Khurana/go-dork-google/issues">
    <img src="https://img.shields.io/github/issues/Abhinandan-Khurana/go-dork-google" alt="Issues">
  </a>
  <a href="https://github.com/Abhinandan-Khurana/go-dork-google/stargazers">
    <img src="https://img.shields.io/github/stars/Abhinandan-Khurana/go-dork-google" alt="Stars">
  </a>
</p>

<p align="center">
  A powerful and flexible Google dorking tool written in Go for efficient subdomain discovery and information gathering
</p>

## 🚀 Features

- 🔍 Advanced Google dorking capabilities
- 🌐 Multiple domain processing
- 🔄 Concurrent searching with customizable threads
- 📊 Multiple output formats (JSON, CSV, TXT)
- 🎯 Targeted subdomain discovery
- 🎨 Colorized output with verbosity controls
- ⚡ Fast and efficient processing
- 🔒 Rate limiting and error handling
- 🎮 User-friendly CLI interface

## 📋 Prerequisites

- Go 1.16 or higher
- Google Custom Search API key
- Google Custom Search Engine ID

## 🛠️ Installation

```bash
# Clone the repository
git clone https://github.com/Abhinandan-Khurana/go-dork-google.git

# Change directory
cd go-dork-google

# Install dependencies
go mod tidy

# Build the binary
go build -o go-dork-google
```

## ⚙️ Configuration

Create a configuration file `google_dorker.yaml` in one of the following locations:

- Current directory
- `~/.config/google_dorker.yaml`
- `/etc/google_dorker.yaml`

```yaml
Google-API:
  - "your-google-api-key-1"
  - "your-google-api-key-2"
Google-CSE-ID:
  - "your-custom-search-engine-id-1"
  - "your-custom-search-engine-id-2"
```

## 🎯 Usage

```bash
# Basic usage
./go-dork-google -d example.com

# Subdomain discovery with JSON output
./go-dork-google -d example.com -subs -format json

# Multiple domain processing
./go-dork-google -d example.com sub1.example.com sub2.example.com -subs

# Custom query with specific output file
./go-dork-google -d example.com -q "password" -o results.csv -format csv

# Silent mode with high concurrency
./go-dork-google -d example.com -silent -concurrent 20
```

### 🎪 Command Line Options

```
Options:
  -q string
        Google dorking query for your target
  -d string
        Target name for Google dorking
  -o string
        File name to save the dorking results
  -format string
        Output format (txt, json, csv) (default "txt")
  -subs
        Only output found subdomains
  -concurrent int
        Number of concurrent searches (default 10)
  -v int
        Verbosity level (0=ERROR, 1=INFO, 2=DEBUG, 3=TRACE) (default 1)
  -version
        Show version information
  -no-color
        Disable color output
  -silent
        Silent mode - only output results
  -timeout duration
        Timeout for the entire search operation (default 5m)
```

## 📋 Example Output

### JSON Format

```json
{
  "example.com": [
    "api.example.com",
    "blog.example.com",
    "dev.example.com",
    "mail.example.com",
    "www.example.com"
  ]
}
```

### CSV Format

```csv
Domain,Subdomain
example.com,api.example.com
example.com,blog.example.com
example.com,dev.example.com
example.com,mail.example.com
example.com,www.example.com
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Google Custom Search API
- The Go Community
- All contributors and users of this tool

## 📞 Contact

Abhinandan Khurana - [@L0u51f3r007](https://x.com/L0u51f3r007)

Project Link: [https://github.com/Abhinandan-Khurana/go-dork-google](https://github.com/Abhinandan-Khurana/go-dork-google)

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/Abhinandan-Khurana">Abhinandan Khurana</a>
</p>
