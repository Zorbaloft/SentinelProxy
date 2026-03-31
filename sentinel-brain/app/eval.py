from typing import Dict, List
from datetime import datetime, timedelta
import re

class RuleEngine:
    def __init__(self, storage):
        self.storage = storage
    
    def evaluate_all_rules(self) -> List[Dict]:
        """Evaluate all active rules and return triggered actions"""
        rules = self.storage.get_active_rules()
        incidents = []
        
        for rule in rules:
            triggered = self._check_rule_triggered(rule)
            if triggered:
                incident = self._create_incident(rule, triggered)
                incidents.append(incident)
                
                # Execute action
                action = rule.get("action", {})
                action_type = action.get("type")
                ttl_sec = action.get("ttlSec", 3600)
                
                if action_type == "block":
                    self.storage.set_blocklist(
                        triggered["ip"],
                        {
                            "reason": action.get("reason", f"rule:{rule.get('name', 'unknown')}"),
                            "createdAt": datetime.utcnow(),
                            "expiresAt": datetime.utcnow() + timedelta(seconds=ttl_sec),
                            "ruleId": str(rule.get("_id", "")) if rule.get("_id") else ""
                        },
                        ttl_sec
                    )
                elif action_type == "redirect":
                    self.storage.set_redirect(
                        triggered["ip"],
                        {
                            "targetUrl": action.get("targetUrl", ""),
                            "reason": action.get("reason", f"rule:{rule.get('name', 'unknown')}"),
                            "createdAt": datetime.utcnow(),
                            "expiresAt": datetime.utcnow() + timedelta(seconds=ttl_sec),
                            "ruleId": str(rule.get("_id", "")) if rule.get("_id") else ""
                        },
                        ttl_sec
                    )
        
        return incidents
    
    def _check_rule_triggered(self, rule: Dict) -> Dict:
        """Check if rule is triggered by scanning Redis counters"""
        conditions = rule.get("conditions", {})
        threshold = rule.get("threshold", 0)
        path_cond = conditions.get("path")
        method = conditions.get("method")
        
        # Build key pattern to scan
        if path_cond:
            path_value = path_cond.get("value", "")
            if method:
                pattern = f"rate:*:{method}:{path_value}"
            else:
                pattern = f"rate:*:{path_value}"
        else:
            if method:
                pattern = f"rate:*:{method}:*"
            else:
                pattern = f"rate:*"
        
        # Scan Redis keys (simplified - use SCAN in production for large datasets)
        triggered_ips = []
        
        # For each matching key, extract IP and check count
        for key_bytes in self.storage.redis.scan_iter(match=pattern.encode()):
            key = key_bytes.decode()
            parts = key.split(":")
            
            if len(parts) < 2:
                continue
            
            ip = parts[1]  # Extract IP from key
            count = self.storage.get_redis_counter(key)
            
            if count >= threshold:
                triggered_ips.append({"ip": ip, "count": count, "key": key})
        
        # Return first triggered IP (or could return all)
        return triggered_ips[0] if triggered_ips else None
    
    def _create_incident(self, rule: Dict, triggered: Dict) -> Dict:
        """Create incident document"""
        action = rule.get("action", {})
        rule_id = rule.get("_id")
        if rule_id:
            # Handle both ObjectId and string
            if hasattr(rule_id, '__str__'):
                rule_id = str(rule_id)
            else:
                rule_id = str(rule_id)
        return {
            "timestamp": datetime.utcnow(),
            "ip": triggered["ip"],
            "ruleId": rule_id or "",
            "ruleName": rule.get("name", "Unknown"),
            "actionTaken": action.get("type", "unknown"),
            "ttlSec": action.get("ttlSec", 3600),
            "evidence": {
                "count": triggered.get("count", 0),
                "threshold": rule.get("threshold", 0)
            },
            "status": "open"
        }
