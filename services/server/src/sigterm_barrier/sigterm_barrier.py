from threading import Condition
import logger 

class SigtermBarrier:
    def __init__(self, quorum: int):
        self.count = 0
        self.quorum = quorum
        self.epoch = 0
        self.sigterm = False 
        self.cvar = Condition()

    def wait(self):
        # Inspirado en la implementacion de std::sync::Barrier de Rust 
        # La barrera es liberada al recibir un sigterm, liberando a los threads en wait
        with self.cvar:
            local_epoch = self.epoch
            self.count += 1 
            if self.count < self.quorum :
                while self.epoch == local_epoch and not self.sigterm:
                    self.cvar.wait()
            else:
                self.epoch += 1 
                self.count = 0
                self.cvar.notify_all()
        logger.info("sigterm-barrier-exit", logger.LogResult.success)

    def sigterm_signal(self):
        with self.cvar:
            self.sigterm = True
            self.cvar.notify_all()
        logger.info("sigterm-barrier-abort", logger.LogResult.in_progress)

