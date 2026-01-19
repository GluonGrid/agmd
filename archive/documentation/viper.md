# Viper - Go Configuration Management

## Package Information

**Import Path**: `github.com/spf13/viper`  
**Latest Version**: v1.21.0  
**License**: MIT  
**Go Get**: `go get github.com/spf13/viper@latest`

## Overview

Viper is a complete configuration solution for Go applications, including 12-Factor apps. It's designed to work within any application and can handle all types of configuration needs and formats.

### Key Features

- Read from JSON, TOML, YAML, HCL, INI, envfile, and Java properties formats
- Set default values for configuration options
- Read from environment variables
- Read from remote config systems (etcd, Consul, Firestore, NATS)
- Read from command line flags (works seamlessly with Cobra/pflag)
- Read from buffers
- Set explicit values
- Live watching and re-reading of config files
- Unmarshaling to structs
- Configuration precedence system

## Configuration Precedence

Viper uses the following precedence order (highest to lowest):

1. **Explicit calls** to `Set`
2. **Flags** (command line)
3. **Environment variables**
4. **Config file**
5. **Key/value store**
6. **Defaults**

## Installation

```bash
go get github.com/spf13/viper@latest
```

## Basic Usage

### Reading a Configuration File

```go
package main

import (
    "fmt"
    "log"
    "github.com/spf13/viper"
)

func main() {
    // Set config name (no extension)
    viper.SetConfigName("config")
    
    // Set config type
    viper.SetConfigType("yaml")
    
    // Add config path(s) to search
    viper.AddConfigPath(".")
    viper.AddConfigPath("$HOME/.myapp")
    viper.AddConfigPath("/etc/myapp/")
    
    // Read config
    err := viper.ReadInConfig()
    if err != nil {
        log.Fatalf("Error reading config: %v", err)
    }
    
    // Get values
    fmt.Println(viper.Get("database.host"))
    fmt.Println(viper.GetString("database.user"))
    fmt.Println(viper.GetInt("database.port"))
}
```

### Example Config File (config.yaml)

```yaml
database:
  host: localhost
  port: 5432
  user: admin
  password: secret

server:
  port: 8080
  debug: true

features:
  - analytics
  - logging
  - monitoring
```

## Setting Defaults

```go
func init() {
    // Set default values
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", 5432)
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.debug", false)
}
```

## Reading Configuration

### Get Methods

```go
// Generic Get (returns interface{})
value := viper.Get("key")

// Type-specific getters
viper.GetBool("server.debug")
viper.GetFloat64("price")
viper.GetInt("server.port")
viper.GetIntSlice("ports")
viper.GetString("database.host")
viper.GetStringMap("database")
viper.GetStringMapString("headers")
viper.GetStringSlice("features")
viper.GetTime("created_at")
viper.GetDuration("timeout")

// Nested access using dot notation
viper.GetString("database.host")
viper.GetInt("database.port")
```

### Check if Key Exists

```go
if viper.IsSet("database.host") {
    host := viper.GetString("database.host")
}
```

### Get All Settings

```go
settings := viper.AllSettings()
// Returns map[string]interface{} of all settings
```

## Environment Variables

### Automatic Environment Binding

```go
// Enable automatic environment variable reading
viper.AutomaticEnv()

// Now any Get() call will also check env vars
port := viper.GetInt("port") // Checks PORT env var
```

### Environment Prefix

```go
viper.SetEnvPrefix("myapp") // Will be uppercased
viper.AutomaticEnv()

// GetString("database") will check MYAPP_DATABASE env var
```

### Environment Key Replacer

```go
import "strings"

// Replace dots with underscores in env vars
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

// viper.Get("database.host") will check DATABASE_HOST
```

### Bind Specific Environment Variables

```go
viper.BindEnv("port") // Binds PORT env var
viper.BindEnv("port", "SERVER_PORT") // Custom env var name
```

## Working with Flags (Cobra Integration)

### Bind to Cobra Flags

```go
package main

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    Run: func(cmd *cobra.Command, args []string) {
        // Access flag values through viper
        port := viper.GetInt("port")
        debug := viper.GetBool("debug")
    },
}

func init() {
    // Define flags
    rootCmd.Flags().IntP("port", "p", 8080, "Server port")
    rootCmd.Flags().Bool("debug", false, "Debug mode")
    
    // Bind flags to viper
    viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
    viper.BindPFlag("debug", rootCmd.Flags().Lookup("debug"))
}

func main() {
    rootCmd.Execute()
}
```

### Bind All Flags

```go
import "github.com/spf13/pflag"

func init() {
    pflag.Int("port", 8080, "Server port")
    pflag.Bool("debug", false, "Debug mode")
    pflag.Parse()
    
    // Bind all pflags
    viper.BindPFlags(pflag.CommandLine)
}
```

## Unmarshaling to Structs

### Basic Unmarshaling

```go
type Config struct {
    Database struct {
        Host     string
        Port     int
        User     string
        Password string
    }
    Server struct {
        Port  int
        Debug bool
    }
}

func main() {
    viper.SetConfigFile("config.yaml")
    viper.ReadInConfig()
    
    var config Config
    err := viper.Unmarshal(&config)
    if err != nil {
        log.Fatalf("Unable to decode into struct: %v", err)
    }
    
    fmt.Println(config.Database.Host)
}
```

### Using Mapstructure Tags

```go
type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"user"`
    Password string `mapstructure:"pass"`
}

type AppConfig struct {
    DB DatabaseConfig `mapstructure:"database"`
}

func main() {
    var config AppConfig
    viper.Unmarshal(&config)
}
```

### Unmarshal Specific Key

```go
var dbConfig DatabaseConfig
viper.UnmarshalKey("database", &dbConfig)
```

## Writing Configuration

### Write to File

```go
// Write current config to preset path
viper.WriteConfig()

// Write to specific path (overwrites)
viper.WriteConfigAs("/path/to/config.yaml")

// Safe write (won't overwrite)
viper.SafeWriteConfig()
viper.SafeWriteConfigAs("/path/to/config.yaml")
```

### Set Values

```go
viper.Set("database.host", "localhost")
viper.Set("database.port", 5432)

// Save to file
viper.WriteConfig()
```

## Watching Configuration Changes

### Watch for Changes

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    fmt.Println("Config file changed:", e.Name)
    // Reload configuration
})

// Config will be automatically reloaded when it changes
```

### Example with Hot Reload

```go
func main() {
    viper.SetConfigFile("config.yaml")
    viper.ReadInConfig()
    
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        log.Println("Config changed, reloading...")
        // Your reload logic here
    })
    
    // Your application continues running
    // Config changes are automatically detected
}
```

## Supported File Formats

### JSON

```json
{
  "database": {
    "host": "localhost",
    "port": 5432
  }
}
```

```go
viper.SetConfigType("json")
```

### YAML

```yaml
database:
  host: localhost
  port: 5432
```

```go
viper.SetConfigType("yaml")
```

### TOML

```toml
[database]
host = "localhost"
port = 5432
```

```go
viper.SetConfigType("toml")
```

### HCL

```hcl
database {
  host = "localhost"
  port = 5432
}
```

```go
viper.SetConfigType("hcl")
```

### INI/Properties

```ini
[database]
host=localhost
port=5432
```

```go
viper.SetConfigType("ini")
```

### ENV/DotEnv

```env
DATABASE_HOST=localhost
DATABASE_PORT=5432
```

```go
viper.SetConfigType("env")
```

## Remote Key/Value Stores

### Etcd

```go
import _ "github.com/spf13/viper/remote"

func main() {
    viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/myapp")
    viper.SetConfigType("json")
    
    err := viper.ReadRemoteConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use viper.Get() as usual
}
```

### Consul

```go
viper.AddRemoteProvider("consul", "127.0.0.1:8500", "config/myapp")
viper.SetConfigType("json")
err := viper.ReadRemoteConfig()
```

### Watch Remote Config

```go
import "time"

func main() {
    viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/myapp")
    viper.SetConfigType("json")
    viper.ReadRemoteConfig()
    
    // Watch for changes
    go func() {
        for {
            time.Sleep(time.Second * 5)
            err := viper.WatchRemoteConfig()
            if err != nil {
                log.Println("Error watching remote config:", err)
                continue
            }
        }
    }()
}
```

## Reading from io.Reader

```go
import "bytes"

configData := []byte(`
database:
  host: localhost
  port: 5432
`)

viper.SetConfigType("yaml")
err := viper.ReadConfig(bytes.NewBuffer(configData))
```

## Advanced Usage

### Multiple Viper Instances

```go
// Create separate instances for different configs
appConfig := viper.New()
appConfig.SetConfigName("app")
appConfig.AddConfigPath(".")
appConfig.ReadInConfig()

dbConfig := viper.New()
dbConfig.SetConfigName("database")
dbConfig.AddConfigPath(".")
dbConfig.ReadInConfig()
```

### Aliases

```go
viper.RegisterAlias("db.host", "database.host")

// Both access the same value
viper.GetString("db.host")
viper.GetString("database.host")
```

### Sub-configurations

```go
database := viper.Sub("database")
if database == nil {
    log.Fatal("database config not found")
}

host := database.GetString("host")
port := database.GetInt("port")
```

## Example: Complete Application

```go
package main

import (
    "fmt"
    "log"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

type Config struct {
    Server struct {
        Port  int    `mapstructure:"port"`
        Host  string `mapstructure:"host"`
        Debug bool   `mapstructure:"debug"`
    } `mapstructure:"server"`
    
    Database struct {
        Host     string `mapstructure:"host"`
        Port     int    `mapstructure:"port"`
        User     string `mapstructure:"user"`
        Password string `mapstructure:"password"`
    } `mapstructure:"database"`
}

var (
    cfgFile string
    config  Config
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Server running on %s:%d\n", 
            config.Server.Host, config.Server.Port)
    },
}

func init() {
    cobra.OnInitialize(initConfig)
    
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", 
        "config file (default is $HOME/.myapp.yaml)")
    rootCmd.Flags().IntP("port", "p", 8080, "server port")
    rootCmd.Flags().Bool("debug", false, "debug mode")
    
    viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port"))
    viper.BindPFlag("server.debug", rootCmd.Flags().Lookup("debug"))
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.AddConfigPath("$HOME")
        viper.AddConfigPath(".")
        viper.SetConfigName(".myapp")
        viper.SetConfigType("yaml")
    }
    
    viper.AutomaticEnv()
    
    // Set defaults
    viper.SetDefault("server.host", "localhost")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.debug", false)
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            log.Println("No config file found, using defaults")
        } else {
            log.Fatalf("Error reading config: %v", err)
        }
    }
    
    if err := viper.Unmarshal(&config); err != nil {
        log.Fatalf("Unable to decode config: %v", err)
    }
    
    log.Printf("Using config file: %s", viper.ConfigFileUsed())
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

## Best Practices

1. **Use defaults** - Always set sensible defaults
2. **Unmarshal to structs** - Type-safe configuration
3. **Environment variables** - Allow env var overrides for 12-factor apps
4. **Config file optional** - App should work without config file
5. **Watch for changes** - Enable hot reload in development
6. **Validate config** - Validate after unmarshaling
7. **Use prefixes** - Use `SetEnvPrefix` to avoid collisions

## Common Patterns for agent-config

Since agent-config is a simple CLI tool that doesn't need complex configuration, we likely won't use Viper. However, if we wanted to add a config file:

```go
// Future enhancement: global config
type AgentConfig struct {
    DefaultSection string   `mapstructure:"default_section"`
    Sections       []string `mapstructure:"sections"`
    SymlinkAll     bool     `mapstructure:"symlink_all"`
}

func init() {
    viper.SetConfigName(".agent-config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("$HOME")
    viper.AddConfigPath(".")
    
    viper.SetDefault("default_section", "General Rules")
    viper.SetDefault("symlink_all", false)
    
    viper.ReadInConfig() // Ignore errors, use defaults
}
```

## Troubleshooting

### Config File Not Found

```go
if err := viper.ReadInConfig(); err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); ok {
        // Config file not found; use defaults
        log.Println("No config file, using defaults")
    } else {
        // Config file found but error reading
        log.Fatalf("Error reading config: %v", err)
    }
}
```

### Concurrent Access

Viper is **not** thread-safe by default. Use sync package for concurrent access:

```go
import "sync"

var (
    mu sync.RWMutex
)

func GetConfig(key string) string {
    mu.RLock()
    defer mu.RUnlock()
    return viper.GetString(key)
}

func SetConfig(key, value string) {
    mu.Lock()
    defer mu.Unlock()
    viper.Set(key, value)
}
```

### Case Sensitivity

Viper is case-insensitive by design. All keys are lowercased:

```go
viper.Set("MyKey", "value")
viper.Get("mykey") // Returns "value"
viper.Get("MYKEY") // Returns "value"
```

## References

- **GitHub**: https://github.com/spf13/viper
- **Go Package**: https://pkg.go.dev/github.com/spf13/viper
- **Examples**: https://github.com/spf13/viper/tree/master/examples
