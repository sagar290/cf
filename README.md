# cf — Cloudflare DNS A Record Updater

`cf` is a lightweight CLI tool written in Go (using Cobra) to quickly update A records (`@` and `www`) for domains managed via the Cloudflare API.

> 🔐 Simple and secure DNS management via terminal.

---

## 🚀 Features

- Update A records for root (`@`) and `www` subdomains
- Automatically enables Cloudflare proxy (orange cloud ☁️)
- Uses secure Cloudflare API Token
- Ideal for deployment and automation scripts

---

## 🛠 Installation

```bash
git clone https://github.com/sagar290/cf.git
cd cf
go build -o cf
sudo mv cf /usr/local/bin/
```

---

## 🔧 Prerequisites

- **Go 1.20+**
- **Cloudflare API Token** with the following permissions:
  - Zone → Zone → Read
  - Zone → DNS → Edit

Set your token in your terminal session:

```bash
export CF_API_TOKEN=your_token_here
```

Or store it in a `.env` file and load it programmatically.

---

## 📦 Usage

### Update both `@` and `www` records:

```bash
cf update dns example.com --ip 1.2.3.4
```

This will:
- Update A record for `example.com`
- Update A record for `www.example.com`
- Proxy both via Cloudflare

### Update a specific record (e.g. only `www`):

```bash
cf update dns example.com www.example.com --ip 1.2.3.4
```

Or root only:

```bash
cf update dns example.com @ --ip 1.2.3.4
```

---

## 📘 Command Structure

```
cf update dns [domain] [fqdn?] --ip <target-ip>
```

- `domain`: root zone (e.g., `example.com`)
- `fqdn` *(optional)*: record to update (`@` or full like `www.example.com`)
- `--ip`: the IP address to assign to the A record

If `fqdn` is not provided, both `@` and `www` will be updated.

---

## 🧪 Example

```bash
cf update dns example.com --ip 1.2.3.4
```

➡️ This updates:
- `example.com` → 1.2.3.4
- `www.example.com` → 1.2.3.4

---

## 🔐 Security

Avoid committing your API token in code.
- Use environment variables (`CF_API_TOKEN`)
- Or secret managers if used in CI/CD

---

## 📄 License

MIT License  
© [Sagar Dash](https://github.com/sagar290)
