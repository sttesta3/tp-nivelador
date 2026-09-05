import os
import sys

import logger
import server
from signal import signal, SIGTERM

SERVER_HOST = os.environ["SERVER_HOST"]
SERVER_PORT = int(os.environ["SERVER_PORT"])
AGENCY_QUORUM_MIN = int(os.environ["AGENCY_QUORUM_MIN"])

def main():
    logger.init()
    s = server.Server(SERVER_HOST, SERVER_PORT, AGENCY_QUORUM_MIN)
    
    def signal_handler(signum,frame): 
        logger.info("server-run", logger.LogResult.in_progress, "SIGTERM. Deteniendo servidor")
        s.stop()
    signal(SIGTERM, signal_handler)

    try:
        s.run()
    except Exception as e:
        logger.error("server-run", logger.LogResult.fail, "err", e)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
