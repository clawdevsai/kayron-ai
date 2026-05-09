"""Thread-safe MT5 SDK adapter"""
import threading
from typing import Optional, Dict, Any, List
from datetime import datetime


class MT5Adapter:
    """Thread-safe wrapper around metatrader5 package"""

    def __init__(self, terminal_path: str = ""):
        self.terminal_path = terminal_path
        self.mt5 = None
        self.lock = threading.RLock()
        self.is_connected = False
        self.connection_time: Optional[datetime] = None

    def initialize(self, login: int, password: str, server: str) -> bool:
        """Initialize MT5 connection (thread-safe)"""
        with self.lock:
            try:
                # Import metatrader5 dynamically
                import metatrader5 as mt5

                self.mt5 = mt5

                # Initialize MT5 connection
                if self.terminal_path:
                    result = mt5.initialize(path=self.terminal_path, login=login, password=password, server=server)
                else:
                    result = mt5.initialize(login=login, password=password, server=server)

                if result:
                    self.is_connected = True
                    self.connection_time = datetime.utcnow()
                    return True
                else:
                    error = mt5.last_error()
                    raise Exception(f"MT5 initialization failed: {error}")

            except ImportError:
                raise ImportError("metatrader5 package not installed")
            except Exception as e:
                self.is_connected = False
                raise e

    def shutdown(self) -> bool:
        """Shutdown MT5 connection (thread-safe)"""
        with self.lock:
            if self.mt5 and self.is_connected:
                try:
                    self.mt5.shutdown()
                    self.is_connected = False
                    return True
                except Exception as e:
                    print(f"Error during MT5 shutdown: {e}")
                    return False
            return True

    def get_account_info(self) -> Optional[Dict[str, Any]]:
        """Get account information (thread-safe)"""
        with self.lock:
            if not self.is_connected or not self.mt5:
                return None

            try:
                account_info = self.mt5.account_info()
                if account_info:
                    return {
                        "login": account_info.login,
                        "balance": account_info.balance,
                        "equity": account_info.equity,
                        "credit": account_info.credit,
                        "margin": account_info.margin,
                        "margin_free": account_info.margin_free,
                        "margin_level": account_info.margin_level,
                        "profit": account_info.profit,
                        "server": account_info.server,
                        "currency": account_info.currency,
                        "leverage": account_info.leverage,
                    }
                return None
            except Exception as e:
                print(f"Error getting account info: {e}")
                return None

    def get_positions(self) -> Optional[List[Dict[str, Any]]]:
        """Get open positions (thread-safe)"""
        with self.lock:
            if not self.is_connected or not self.mt5:
                return None

            try:
                positions = self.mt5.positions_get()
                if positions:
                    return [
                        {
                            "ticket": p.ticket,
                            "symbol": p.symbol,
                            "magic": p.magic,
                            "identifier": p.identifier,
                            "type": p.type,
                            "volume": p.volume,
                            "open_price": p.price_open,
                            "current_price": p.price_current,
                            "stop_loss": p.sl,
                            "take_profit": p.tp,
                            "profit": p.profit,
                            "time_open": p.time,
                            "comment": p.comment,
                        }
                        for p in positions
                    ]
                return []
            except Exception as e:
                print(f"Error getting positions: {e}")
                return None

    def place_order(
        self,
        symbol: str,
        order_type: int,
        volume: float,
        price: float,
        stop_loss: Optional[float] = None,
        take_profit: Optional[float] = None,
        comment: str = "",
    ) -> Optional[Dict[str, Any]]:
        """Place order (thread-safe)"""
        with self.lock:
            if not self.is_connected or not self.mt5:
                return None

            try:
                request = {
                    "action": self.mt5.TRADE_ACTION_DEAL,
                    "symbol": symbol,
                    "volume": volume,
                    "type": order_type,
                    "price": price,
                    "comment": comment,
                }

                if stop_loss:
                    request["sl"] = stop_loss
                if take_profit:
                    request["tp"] = take_profit

                result = self.mt5.order_send(request)

                if result and result.retcode == self.mt5.TRADE_RETCODE_DONE:
                    return {
                        "order_id": result.deal,
                        "ticket": result.ticket,
                        "status": "SUCCESS",
                        "volume": result.volume,
                        "price": result.price,
                    }
                else:
                    return {"status": "FAILED", "retcode": result.retcode if result else None}

            except Exception as e:
                print(f"Error placing order: {e}")
                return None

    def close_position(self, ticket: int, volume: Optional[float] = None) -> Optional[Dict[str, Any]]:
        """Close position (thread-safe)"""
        with self.lock:
            if not self.is_connected or not self.mt5:
                return None

            try:
                position = self.mt5.position_get(ticket=ticket)
                if not position:
                    return {"status": "FAILED", "reason": "Position not found"}

                request = {
                    "action": self.mt5.TRADE_ACTION_DEAL,
                    "symbol": position.symbol,
                    "type": self.mt5.ORDER_TYPE_SELL if position.type == 0 else self.mt5.ORDER_TYPE_BUY,
                    "volume": volume or position.volume,
                    "price": self.mt5.symbol_info_tick(position.symbol).bid,
                    "position": ticket,
                }

                result = self.mt5.order_send(request)

                if result and result.retcode == self.mt5.TRADE_RETCODE_DONE:
                    return {
                        "order_id": result.deal,
                        "status": "SUCCESS",
                        "closed_volume": result.volume,
                    }
                else:
                    return {"status": "FAILED", "retcode": result.retcode if result else None}

            except Exception as e:
                print(f"Error closing position: {e}")
                return None

    def get_symbol_info(self, symbol: str) -> Optional[Dict[str, Any]]:
        """Get symbol information (thread-safe)"""
        with self.lock:
            if not self.is_connected or not self.mt5:
                return None

            try:
                info = self.mt5.symbol_info(symbol)
                if info:
                    return {
                        "symbol": info.name,
                        "bid": info.bid,
                        "ask": info.ask,
                        "spread": info.ask - info.bid,
                        "point": info.point,
                        "digits": info.digits,
                        "volume_min": info.volume_min,
                        "volume_max": info.volume_max,
                        "volume_step": info.volume_step,
                    }
                return None
            except Exception as e:
                print(f"Error getting symbol info: {e}")
                return None

    def health_check(self) -> Dict[str, Any]:
        """Health check of MT5 connection"""
        with self.lock:
            return {
                "is_connected": self.is_connected,
                "connection_time": self.connection_time.isoformat() if self.connection_time else None,
                "uptime_seconds": (datetime.utcnow() - self.connection_time).total_seconds()
                if self.connection_time
                else None,
            }
