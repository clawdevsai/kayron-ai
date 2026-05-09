from setuptools import setup, find_packages

setup(
    name="mt5-grpc-mcp",
    version="0.1.0",
    description="Python gRPC MCP server for MT5 operations",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    python_requires=">=3.8",
    install_requires=[
        "grpcio==1.60.0",
        "grpcio-tools==1.60.0",
        "metatrader5==5.0.45",
        "pydantic==2.5.0",
        "sqlalchemy==2.0.23",
        "python-dotenv==1.0.0",
    ],
    extras_require={
        "dev": [
            "pytest==7.4.3",
            "pytest-asyncio==0.21.1",
        ],
    },
    entry_points={
        "console_scripts": [
            "mt5-grpc-server=server:main",
        ],
    },
)
