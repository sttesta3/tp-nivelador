import socket
import logger
import safe_socket
import lottery 
import threading
from sigterm_barrier import SigtermBarrier 

shutdown_event = threading.Event()

class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.agency_quorum_min = agency_quorum_min
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.bind((self.server_host,self.server_port))        
        self.socket.listen()        

    def _format_response(self, bet: Lottery.Bet) -> bytes:
        # Saca la informacion de la agencia del mensaje, tal que siga respetando el echo server        
        response = f"{bet.first_name},{bet.last_name},{str(bet.document)},{bet.birthdate},{str(bet.number)}"
        response_bytes = response.encode('utf-8')
        frame_len = len(response_bytes).to_bytes(2, byteorder='big')
        return frame_len + response_bytes

    def _format_ack(self) -> bytes:
        return (0).to_bytes(2, byteorder='big')

    def _process_message(self, message, agency_id) -> [lottery.Bet]:
        bets = []
        for message_line in message.decode('utf-8').split('\n'):
            if message_line: 
                message_fields = message_line.split(',')
                if len(message_fields) != 5:
                    logger.error("process-message","Mensaje mal formateado",message_lines)
                    raise Exception()

                first_name = message_fields[0]
                last_name = message_fields[1]
                document = int(message_fields[2])
                birthdate = message_fields[3]
                number = int(message_fields[4])

                bets.append(lottery.Bet(agency_id,first_name,last_name,document,birthdate,number))

        return bets

    def _receive_message(self, client_socket):
        message = bytearray()
        message_len_bytes = safe_socket.recv_all(client_socket, 2)
        if message_len_bytes:
            message_len = int.from_bytes(message_len_bytes, byteorder='big')
            if message_len > 0:
                message = safe_socket.recv_all(client_socket, message_len)
        return message

    def _handle_client(self, client_socket, sync_barrier: threading.Barrier):
        action = "handle-client"

        # Lottery creation
        agency_id = int(self._receive_message(client_socket))
        client_lottery_file = str(agency_id) 
        with open(client_lottery_file, "w"):
            client_lottery =  lottery.Lottery(client_lottery_file)
        safe_socket.send_all(client_socket, self._format_ack())

        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            while not shutdown_event.is_set():
                logger.info(action, logger.LogResult.in_progress, "messages-amount", message_amount)
                client_message = self._receive_message(client_socket)
                if not client_message:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )

                    # Debemos tener AGENCY_QUORUM_MIN para ejecutar la siguiente seccion
                    sync_barrier.wait()

                    winners = []
                    if not shutdown_event.is_set():
                        winners = [bet for bet in client_lottery.load_bets() if client_lottery.has_won(bet)] 
                    if not shutdown_event.is_set():
                        # Llegado a este punto es mejor que el hilo termine.
                        # Es una loteria, los ganadores se cuentan por unidad 
                        for winner in winners:
                            safe_socket.send_all(client_socket, self._format_response(winner))

                    client_socket.close()
                    
                    return
                
                bets = self._process_message(client_message, agency_id)
                client_lottery.store_bets(bets)
                message_amount += 1

                safe_socket.send_all(client_socket, self._format_ack())
            client_socket.close()

        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            if client_socket:
                client_socket.close()
            if not shutdown_event.is_set():
                raise e


    def run(self):
        action = "accept-connection"
        sync_barrier = SigtermBarrier(self.agency_quorum_min)
        threads = []
        while not shutdown_event.is_set():
            try:
                logger.info(action, logger.LogResult.in_progress)
                client_socket, _ = self.socket.accept()   # TODO Gemini here is the issue ? 
                logger.info(action, logger.LogResult.success)
                client_thread = threading.Thread(
                    target=self._handle_client,
                    args=(client_socket,sync_barrier,)
                )
                client_thread.start()                    
                threads.append(client_thread)
                # Eliminamos los threads que finalizaron
                threads = [thread for thread in threads if thread.is_alive()]
            except Exception as e:
                if shutdown_event.is_set():
                    break
                logger.error(action, logger.LogResult.fail)
                raise e
            
        logger.info("SIGTERM. Deteniendo servidor", logger.LogResult.in_progress)
        sync_barrier.sigterm_signal()
        for thread in threads:
            logger.info(f"Deteniendo hilo {thread.native_id}", logger.LogResult.in_progress)
            thread.join()
            logger.info(f"SIGTERM. Hilo detenido {thread.native_id}", logger.LogResult.in_progress)

    def stop(self):
        shutdown_event.set()
        try:
            self.socket.close()
        except Exception as e:
            logger.error("sigterm-stop", logger.LogResult.fail, e)
            raise e
