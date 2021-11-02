#import board
#import neopixel
import threading
import socket
import select

PORT = 2023

bindsocket = socket.socket()
bindsocket.bind(('', PORT))
bindsocket.listen(5)
print("listening...")

def clientInputLoop(sock, fromaddr):
    #NUM_LEDS = 512
    #BRIGHTNESS = 0.1
    #pixels = neopixel.NeoPixel(board.D18, NUM_LEDS, brightness=BRIGHTNESS, auto_write=False)

    while True:
        try:
            ready_to_read, _, _ = select.select([sock,], [sock,], [], 5)
            if len(ready_to_read) > 0:
                #read led data from the socket
                clientData = sock.recv(24).decode('utf-8').strip() #expected format: 'R,G,B[,Brightness]'
                if clientData == '':
                    break
                dataSplit = clientData.split(',')
                #if the brightness was sent, reset the pixel brightness
                # if len(dataSplit) == 4:
                #     print(pixels.brightness)
                #     pixels.brightness = float(dataSplit[3])
                #set all led color
                print(fromaddr, '->',(int(dataSplit[0]), int(dataSplit[1]), int(dataSplit[2])))
                # pixels.fill((int(dataSplit[0]), int(dataSplit[1]), int(dataSplit[2])))
                # #update pixels
                # pixels.show()
        except:
            sock.shutdown(2)    # 0 = done receiving, 1 = done sending, 2 = both1
            sock.close()
            # connection error event here, maybe reconnect
            print('connection error')
            return

while True:
    newsocket, fromaddr = bindsocket.accept()
    print('Connection from:', fromaddr)
    t1 = threading.Thread(target=clientInputLoop, args=(newsocket,fromaddr,))
    t1.start()
