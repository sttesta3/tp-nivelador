from threading import Condition

class SigtermBarrier:
    def __init__(self, quorum: int):
        self.count = 0
        self.quorum = quorum
        self.epoch = 0
        self.sigterm = False 
        self.cvar = Condition()

    def wait(self):
        # Inspirado en la implementacion de std::sync::Barrier de Rust 
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

    def sigterm_signal(self):
        with self.cvar:
            self.sigterm = True
            self.cvar.notify_all()
