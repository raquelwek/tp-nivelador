import threading
from contextlib import contextmanager

class RWLock:
    def __init__(self):
        self.lock = threading.Lock()
        self.condition = threading.Condition(self.lock)
        self.readers = 0
        self.writing = False
    '''
    Adquiere el lock de lectura. Si hay un escritor activo, espera de forma
    bloqueante hasta que este finalice. Una vez adquirido, incrementa el
    contador de lectores.
    '''
    def acquire_read(self):
        with self.lock:
            while self.writing:
                self.condition.wait()
            self.readers += 1
    '''
    Libera el acceso de lectura decrementando el contador de lectores.
    Si no quedan lectores activos, notifica a los threads que están
    esperando para adquirir el lock.
    '''
    def release_read(self):
        with self.lock:
            self.readers -= 1
            if self.readers == 0:
                self.condition.notify_all()

    '''
    Adquiere el lock de escritura. Espera de forma bloqueante mientras
    haya otro escritor activo o lectores activos. De esta forma, cuando
    adquiere el lock, el thread es el único que puede acceder al archivo,
    garantizando la exclusión mutua durante la escritura.
    '''
    def acquire_write(self):
        with self.lock:
            while self.writing or self.readers > 0:
                self.condition.wait()
            self.writing = True
    '''
    Libera el lock de escritura y notifica a los threads que están
    esperando para adquirir el lock, permitiendo que vuelvan a evaluar
    si pueden acceder al archivo.
    '''
    def release_write(self):
        with self.lock:
            self.writing = False
            self.condition.notify_all()

## -- funciones para adquirir el lock usando with , e.g: with <read_lock>: 
    @contextmanager
    def read_lock(self):
        self.acquire_read()
        try:
            yield
        finally:
            self.release_read()

    @contextmanager
    def write_lock(self):
        self.acquire_write()
        try:
            yield
        finally:
            self.release_write()
