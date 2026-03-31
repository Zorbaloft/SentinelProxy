import os
import time
import logging
from app.storage import Storage
from app.eval import RuleEngine
from app.ai import AIHeuristics

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def main():
    poll_interval = int(os.getenv("POLL_INTERVAL_SEC", "5"))
    
    logger.info("Initializing Sentinel Brain...")
    storage = Storage()
    rule_engine = RuleEngine(storage)
    ai = AIHeuristics(storage)
    
    logger.info(f"Starting rule evaluation loop (poll interval: {poll_interval}s)")
    
    while True:
        try:
            # Evaluate all rules
            incidents = rule_engine.evaluate_all_rules()
            
            # Save incidents
            for incident in incidents:
                storage.create_incident(incident)
                logger.info(f"Created incident: {incident.get('ruleName')} for IP {incident.get('ip')}")
            
            # AI analysis (optional)
            if ai.should_auto_block():
                recommendations = ai.analyze_anomalies()
                # Process recommendations...
            
            time.sleep(poll_interval)
            
        except Exception as e:
            logger.error(f"Error in evaluation loop: {e}", exc_info=True)
            time.sleep(poll_interval)

if __name__ == "__main__":
    main()
