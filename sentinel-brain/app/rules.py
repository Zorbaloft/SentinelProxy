import re
from typing import Dict, List, Optional

class RuleEvaluator:
    def __init__(self, storage):
        self.storage = storage
    
    def evaluate_rule(self, rule: Dict, ip: str) -> bool:
        """Evaluate if a rule matches for a given IP"""
        conditions = rule.get("conditions", {})
        threshold = rule.get("threshold", 0)
        window_sec = rule.get("windowSec", 60)
        
        # Check path condition
        path_cond = conditions.get("path")
        if path_cond:
            if not self._match_path(path_cond, ""):  # Path will be checked per counter
                return False
        
        # Check method condition
        method = conditions.get("method")
        
        # Check user agent condition (would need to check logs, skip for now)
        # user_agent = conditions.get("userAgent")
        
        # Check status condition
        status_cond = conditions.get("status")
        
        # Build Redis keys to check
        keys_to_check = []
        
        if path_cond:
            path_value = path_cond.get("value", "")
            if method:
                keys_to_check.append(f"rate:{ip}:{method}:{path_value}")
            keys_to_check.append(f"rate:{ip}:{path_value}")
            
            # Status-based counters
            if status_cond:
                status_class = self._get_status_class(status_cond)
                if status_class:
                    keys_to_check.append(f"rate_status:{ip}:{path_value}:{status_class}")
        else:
            # No path condition, check all paths for this IP
            # This is simplified - in production would need to scan keys
            if method:
                keys_to_check.append(f"rate:{ip}:{method}:*")
            keys_to_check.append(f"rate:{ip}:*")
        
        # Check if any counter exceeds threshold
        for key in keys_to_check:
            # Handle wildcards (simplified - would need SCAN in production)
            if "*" in key:
                continue  # Skip wildcard for now
            
            count = self.storage.get_redis_counter(key)
            if count >= threshold:
                return True
        
        return False
    
    def _match_path(self, path_cond: Dict, path: str) -> bool:
        """Match path against condition"""
        cond_type = path_cond.get("type", "exact")
        value = path_cond.get("value", "")
        
        if cond_type == "exact":
            return path == value
        elif cond_type == "prefix":
            return path.startswith(value)
        elif cond_type == "regex":
            try:
                return bool(re.match(value, path))
            except:
                return False
        return False
    
    def _get_status_class(self, status_cond: Dict) -> Optional[str]:
        """Get status class from condition"""
        cond_type = status_cond.get("type")
        
        if cond_type == "range":
            min_val = status_cond.get("min", 0)
            max_val = status_cond.get("max", 999)
            # Determine class from range
            if 200 <= min_val <= max_val < 300:
                return "2xx"
            elif 300 <= min_val <= max_val < 400:
                return "3xx"
            elif 400 <= min_val <= max_val < 500:
                return "4xx"
            elif 500 <= min_val <= max_val < 600:
                return "5xx"
        elif cond_type == "class":
            return status_cond.get("value", "").lower()
        elif cond_type == "exact":
            status_code = status_cond.get("value", 0)
            if 200 <= status_code < 300:
                return "2xx"
            elif 300 <= status_code < 400:
                return "3xx"
            elif 400 <= status_code < 500:
                return "4xx"
            elif 500 <= status_code < 600:
                return "5xx"
        
        return None
