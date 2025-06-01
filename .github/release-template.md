# Traffic Control Go Release Template

## 🎯 **This Release Includes**

### 🚀 **New Features**
- List major new features here
- Use bullet points for clarity

### 🐛 **Bug Fixes**
- Fixed issues and bugs
- Reference GitHub issue numbers when possible

### 📈 **Improvements**
- Performance improvements
- Code quality enhancements
- Documentation updates

### ⚠️ **Breaking Changes**
- Any breaking API changes
- Migration instructions if needed

## 🛠️ **Installation**

### Quick Install (Linux/macOS)
```bash
# Download and install
curl -L https://github.com/RNG999/traffic-control-go/releases/latest/download/traffic-control-go_linux_amd64.tar.gz | tar -xz
sudo cp traffic-control /usr/local/bin/
```

### Windows
```powershell
# Download zip from releases page and extract
# Add to PATH for system-wide access
```

## 📋 **Usage Examples**

```bash
# Simple rate limiting with TBF
sudo traffic-control tbf eth0 1:0 100Mbps

# Priority-based scheduling with PRIO
sudo traffic-control prio eth0 1:0 3

# Fair queuing with FQ_CODEL
sudo traffic-control fq_codel eth0 1:0 --target 1000 --ecn

# Complex HTB hierarchy
sudo traffic-control htb eth0 1:0 1:999 \
    --class 1:10,parent=1:,rate=60Mbps,ceil=80Mbps \
    --class 1:20,parent=1:,rate=30Mbps,ceil=50Mbps
```

## 🔧 **What's Changed**
<!-- This will be automatically filled by GoReleaser -->

## 📊 **Assets**
<!-- GoReleaser will automatically attach build artifacts -->

---

**Full Changelog**: https://github.com/RNG999/traffic-control-go/compare/PREVIOUS_TAG...CURRENT_TAG