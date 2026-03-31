import os
from pymongo import MongoClient
import redis
from typing import List, Dict, Optional

class Storage:
    def __init__(self):
        mongo_uri = os.getenv("MONGO_URI", "mongodb://localhost:27017/sentinel")
        redis_addr = os.getenv("REDIS_ADDR", "localhost:6379")
        
        self.mongo = MongoClient(mongo_uri)
        self.db = self.mongo.sentinel
        self.redis = redis.Redis.from_url(f"redis://{redis_addr}", decode_responses=False)
    
    def get_active_rules(self) -> List[Dict]:
        """Get all enabled rules from MongoDB"""
        return list(self.db.rules.find({"enabled": True}))
    
    def get_redis_counter(self, key: str) -> int:
        """Get Redis counter value"""
        val = self.redis.get(key)
        return int(val) if val else 0
    
    def set_blocklist(self, ip: str, data: Dict, ttl_sec: int):
        """Set blocklist entry in Redis"""
        import json
        from datetime import datetime
        # Convert datetime objects to ISO strings
        data_copy = {}
        for k, v in data.items():
            if isinstance(v, datetime):
                data_copy[k] = v.isoformat()
            else:
                data_copy[k] = v
        key = f"blocklist:{ip}"
        self.redis.setex(key, ttl_sec, json.dumps(data_copy))
    
    def set_redirect(self, ip: str, data: Dict, ttl_sec: int):
        """Set redirect entry in Redis"""
        import json
        from datetime import datetime
        # Convert datetime objects to ISO strings
        data_copy = {}
        for k, v in data.items():
            if isinstance(v, datetime):
                data_copy[k] = v.isoformat()
            else:
                data_copy[k] = v
        key = f"redirect:{ip}"
        self.redis.setex(key, ttl_sec, json.dumps(data_copy))
    
    def create_incident(self, incident: Dict):
        """Create incident in MongoDB"""
        self.db.incidents.insert_one(incident)
