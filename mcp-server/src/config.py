"""Configuration loader for MT5 gRPC service"""
import os
from typing import Optional, Dict, Any
from pydantic import BaseModel, Field, validator


class MT5Settings(BaseModel):
    """MT5 terminal connection settings"""
    terminal_path: str = Field(default="C:\\Program Files\\MetaTrader 5\\terminal64.exe")
    login: int
    password: str
    server: str = "MetaQuotes-Demo"

    class Config:
        case_sensitive = False


class ServerSettings(BaseModel):
    """gRPC server settings"""
    host: str = Field(default="0.0.0.0")
    port: int = Field(default=50051)
    workers: int = Field(default=10)

    @validator("port")
    def port_must_be_valid(cls, v):
        if v < 1 or v > 65535:
            raise ValueError("Port must be between 1 and 65535")
        return v


class DatabaseSettings(BaseModel):
    """Database settings"""
    db_path: str = Field(default="mcp-server.db")
    enable_persistence: bool = Field(default=True)
    connection_timeout: int = Field(default=5)


class LoggingSettings(BaseModel):
    """Logging configuration"""
    level: str = Field(default="INFO")
    format: str = Field(default="json")
    output: str = Field(default="file")
    log_file: str = Field(default="mcp-server.log")


class Config(BaseModel):
    """Main configuration"""
    mt5: MT5Settings
    server: ServerSettings = Field(default_factory=ServerSettings)
    database: DatabaseSettings = Field(default_factory=DatabaseSettings)
    logging: LoggingSettings = Field(default_factory=LoggingSettings)
    api_keys: Dict[str, str] = Field(default_factory=dict)

    class Config:
        case_sensitive = False


def load_config(config_path: Optional[str] = None) -> Config:
    """Load configuration from environment or config file"""
    import json
    import yaml

    # Default config path
    if config_path is None:
        config_path = os.getenv("MT5_CONFIG_PATH", "config.yaml")

    config_data = {}

    # Load from file if exists
    if os.path.exists(config_path):
        with open(config_path, "r") as f:
            if config_path.endswith(".yaml") or config_path.endswith(".yml"):
                try:
                    import yaml
                    config_data = yaml.safe_load(f) or {}
                except ImportError:
                    print("Warning: PyYAML not installed, skipping YAML config")
            elif config_path.endswith(".json"):
                config_data = json.load(f)

    # Override with environment variables
    if os.getenv("MT5_LOGIN"):
        if "mt5" not in config_data:
            config_data["mt5"] = {}
        config_data["mt5"]["login"] = int(os.getenv("MT5_LOGIN"))

    if os.getenv("MT5_PASSWORD"):
        if "mt5" not in config_data:
            config_data["mt5"] = {}
        config_data["mt5"]["password"] = os.getenv("MT5_PASSWORD")

    if os.getenv("MT5_SERVER"):
        if "mt5" not in config_data:
            config_data["mt5"] = {}
        config_data["mt5"]["server"] = os.getenv("MT5_SERVER")

    if os.getenv("GRPC_PORT"):
        if "server" not in config_data:
            config_data["server"] = {}
        config_data["server"]["port"] = int(os.getenv("GRPC_PORT"))

    if os.getenv("DB_PATH"):
        if "database" not in config_data:
            config_data["database"] = {}
        config_data["database"]["db_path"] = os.getenv("DB_PATH")

    # Validate required fields
    if "mt5" not in config_data or "login" not in config_data.get("mt5", {}):
        raise ValueError("MT5_LOGIN environment variable or config.yaml mt5.login required")
    if "mt5" not in config_data or "password" not in config_data.get("mt5", {}):
        raise ValueError("MT5_PASSWORD environment variable or config.yaml mt5.password required")

    return Config(**config_data)


def validate_config(config: Config) -> bool:
    """Validate configuration settings"""
    errors = []

    # Validate MT5 settings
    if not config.mt5.login or config.mt5.login <= 0:
        errors.append("Valid MT5 login required")
    if not config.mt5.password:
        errors.append("MT5 password required")
    if not config.mt5.server:
        errors.append("MT5 server required")

    # Validate server settings
    if config.server.port <= 0:
        errors.append("Valid server port required")

    if errors:
        raise ValueError(f"Configuration validation failed: {'; '.join(errors)}")

    return True
