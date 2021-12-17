### If you dont have an led strip then use this(Test_setLeds.py) instead of setLeds.py
import threading
import socket
import select

PORT = 4444

bindsocket = socket.socket()
bindsocket.bind(('', PORT))
bindsocket.listen(5)
print("listening...")

def clientInputLoop(sock, fromaddr):
    while True:
        try:
            #read led data from the socket
            clientData = sock.recv(128).decode('utf-8').strip()
            if clientData == '':
                break
            print(fromaddr, '->', clientData)
        except Exception as e:
            sock.shutdown(2)    # 0 = done receiving, 1 = done sending, 2 = both1
            sock.close()
            # connection error event here, maybe reconnect
            print('connection error:', e)
            return

while True:
    newsocket, fromaddr = bindsocket.accept()
    print('Connection from:', fromaddr)
    t1 = threading.Thread(target=clientInputLoop, args=(newsocket,fromaddr,))
    t1.start()
