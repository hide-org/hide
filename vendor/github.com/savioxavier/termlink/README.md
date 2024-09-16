<!-- markdownlint-disable MD010 MD033 MD001 -->

# termlink

> Clickable links in the terminal for Go

[![Release](https://img.shields.io/github/release/savioxavier/termlink.svg)](https://github.com/savioxavier/termlink/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/savioxavier/termlink.svg)](https://pkg.go.dev/github.com/savioxavier/termlink)
[![License](https://img.shields.io/github/license/savioxavier/termlink)](/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/savioxavier/termlink)](https://goreportcard.com/report/github.com/savioxavier/termlink)

![termlink_demo_nyoom](https://user-images.githubusercontent.com/38729705/163217599-6fecf944-c10e-4546-9669-1c7d047da55e.gif) <sup>*</sup>

**Termlink is a Go package that allows you to create fully customizable clickable links in the terminal. It is the Go version of Sindre Sorhus' popular [terminal-link](https://github.com/sindresorhus/terminal-link/) library.**

**It includes multiple features including dynamic and fully customizable colored links in terminal, along with smart hyperlink support detection.**

## 🛠️ Install

Using `go get`:

```text
go get github.com/savioxavier/termlink
```

---

## 🔗 Usage

- Basic regular link:

```go
import (
	"fmt"

	"github.com/savioxavier/termlink"
)

func main() {
	fmt.Println(termlink.Link("Example", "https://example.com"))
}
```

- Customizable colored link:

```go
import (
	"fmt"

	"github.com/savioxavier/termlink"
)

func main() {
	fmt.Println(termlink.ColorLink("Example", "https://example.com", "italic green"))
}
```

- Check if your current terminal supports hyperlinks:

```go
import (
	"fmt"

	"github.com/savioxavier/termlink"
)

func main() {
	fmt.Println(termlink.SupportsHyperlinks())
	// prints either true or false, depending on your terminal
}
```

- Since `termlink.Link` returns a string, you can use it with your favorite text color formatting libraries such as [fatih/color](https://github.com/fatih/color), [mgutz/ansi](https://github.com/mgutz/ansi), etc. Alternatively, you can use `termlink.ColorLink` as well.

```go
import (
	"fmt"

	"github.com/fatih/color"
	"github.com/savioxavier/termlink"
)

func main() {
	// With fatih/color package
	color.Cyan(termlink.Link("Example link using the fatih/color package", "https://example.com"))
}
```

> **Note**
> **For unsupported terminals, the link will be printed in parentheses after the text (see below image)**
>
> ![image](https://user-images.githubusercontent.com/38729705/163216009-abb81d39-aff0-4fb5-8c5f-da36e241b395.png)

---

## 🍵 Examples

More examples can be found in the [`examples/`](examples/) directory.

For a quick demo, execute the following commands in your terminal:

```bash
git clone https://github.com/savioxavier/termlink.git
cd termlink/
go get github.com/savioxavier/termlink
go run examples/start.go
```

---

## 🔮 Features

- **`termlink.Link(text, url, [shouldForce])`**

  - Creates a regular, clickable link in the terminal
  - For unsupported terminals, the link will be printed in parentheses after the text: `Example Link (https://example.com)`.
  - The `shouldForce` is an optional boolean parameter which allows you to force the above unsupported terminal hyperlinks format `text (url)` to be printed, even in supported terminals

- **`termlink.ColorLink(text, url, color, [shouldForce])`**

  - Creates a clickable link in the terminal with custom color formatting
  - Examples of color options include:
    - Foreground only: `green`, `red`, `blue`, etc.
    - Background only: `bgGreen`, `bgRed`, `bgBlue`, etc.
    - Foreground and background: `green bgRed`, `bgBlue red`, etc.
    - With formatting: `green bold`, `red bgGreen italic`, `italic blue bgGreen`, etc.
  - The `shouldForce` is an optional boolean parameter which allows you to force the above unsupported terminal hyperlinks format `text (url)` to be printed, even in supported terminals

> `shouldForce` can be used in the following manner:
>
> ```go
> termlink.Link(text, url, true)
> termlink.ColorLink(text, url, color, true)
> ```
>
> You don't always need to specify this argument. By default, this parameter is `false`

- **`termlink.SupportsHyperlinks()`**:

  - Returns `true` if the terminal supports hyperlinks, `false` otherwise.

---

## 🧪 Tests

You can run unit tests _locally_ by running the following command:

```bash
go test -v
```

Tests can be found in [`termlink_test.go`](./termlink_test.go)

---

## ❤️ Support

You can support further development of this project by **giving it a 🌟** and help me make even better stuff in the future by **buying me a ☕**

<a href="https://www.buymeacoffee.com/savioxavier">
<img src="https://cdn.buymeacoffee.com/buttons/v2/default-blue.png" height="50px">
</a>

<br>

**Also, if you liked this repo, consider checking out my other projects, that would be real cool!**

---

## 💫 Attributions and special thanks

- [terminal-link](https://github.com/sindresorhus/terminal-link) - Sindre Sorhus' original package for providing inspiration for this package.
- [supports-hyperlinks](https://github.com/zkat/supports-hyperlinks) - Zkat's package for additional hyperlink handling support.

<sub><sup>* The paperclip icon shown in the demo at the top of this README isn't included when you create the link, it's purely for decorative purposes only.</sup></sub>
