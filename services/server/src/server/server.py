import socket
import logger
import safe_socket
import lottery 

_ECHO_SERVER_MESSAGE_SIZE = 1024

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port

    def _format_response(self, bet: Lottery.bet) -> bytes:
        # Saca la informacion de la agencia del mensaje, tal que siga respetando el echo server        
        response = f"{bet.first_name},{bet.last_name},{str(bet.document)},{bet.birthdate},{str(bet.number)}"
        response_bytes = response.encode('utf-8')
        return bytes([len(response_bytes)]) + response_bytes

    def _format_ack(self) -> bytes:
        return bytes([0])

    def _process_message(self, message, agency_id) -> [lottery.Bet]:
        message_fields = message.decode('utf-8').split(',')
        if len(message_fields) != 5:
            logger.Warn("process-message",logger.Fail, "Mensaje mal formateado, debe tener cinco campos separados por coma")
            raise Exception()

        first_name = message_fields[0]
        last_name = message_fields[1]
        document = int(message_fields[2])
        birthdate = message_fields[3]
        number = int(message_fields[4])

        return [lottery.Bet(agency_id,first_name,last_name,document,birthdate,number)]        

    def _receive_message(self, client_socket):
        message = bytearray()
        message_len = safe_socket.recv_all(client_socket, 1)
        if message_len:
            message = safe_socket.recv_all(client_socket, message_len[0])
        return message

    def _handle_client(self, client_socket):
        action = "handle-client"

        agency_id = self._receive_message(client_socket)
        client_lottery =  lottery.Lottery(str(agency_id))
        safe_socket.send_all(client_socket, self._format_ack())

        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                client_message = self._receive_message(client_socket)
                if not client_message:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )
                    for bet in lottery.load_bets():
                        if lottery.has_won(bet):
                            safe_socket.send_all(client_socket, self._format_response(winner))
                    return
                
                bets = self._process_message(client_message, agency_id)
                client_lottery.store_bets(bets)
                message_amount += 1

                safe_socket.send_all(client_socket, self._format_ack())
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                    with client_socket:
                        logger.info(action, logger.LogResult.success)
                        self._handle_client(client_socket)
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e