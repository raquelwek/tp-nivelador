import threading

class RWLock:
    def __init__(self):
        self.lock = threading.Lock()
        self.condition = threading.Condition(self.lock)
        self.readers = 0
        self.writing = False

    def acquire_read(self):
        with self.lock:
            while self.writing:
                self.condition.wait()
            self.readers += 1

    def release_read(self):
        with self.lock:
            self.readers -= 1
            if self.readers == 0:
                self.condition.notify_all()

    def acquire_write(self):
        with self.lock:
            while self.writing or self.readers > 0:
                self.condition.wait()
            self.writing = True

    def release_write(self):
        with self.lock:
            self.writing = False
            self.condition.notify_all()
