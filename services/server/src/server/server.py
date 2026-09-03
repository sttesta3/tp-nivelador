import socket
import logger
import safe_socket
import lottery 

_ECHO_SERVER_MESSAGE_SIZE = 1024

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = lottery.Lottery('output/server-output.csv')

    def _format_response(self, bet: Lottery.bet):
        # Saca la informacion de la agencia del mensaje, tal que siga respetando el echo server        
        response = f"{bet.first_name},{bet.last_name},{str(bet.document)},{bet.birthdate},{str(bet.number)}"
        response_bytes = response.encode('utf-8')
        return bytes([len(response_bytes)]) + response_bytes

    def _process_message(self, message) -> [lottery.Bet]:

        message_len = message[0]
        agency_len = message[1]
        agencyId = int(message[2:2+agency_len].decode('utf-8'))
        message_fields = message[2+agency_len:].decode('utf-8').split(',')
        if len(message_fields) != 5:
            logger.Warn("process-message",logger.Fail, "Mensaje mal formateado, debe tener cinco campos separados por coma")
            raise Exception()

        first_name = message_fields[0]
        last_name = message_fields[1]
        document = int(message_fields[2])
        birthdate = message_fields[3]
        number = int(message_fields[4])

        return [lottery.Bet(agencyId,first_name,last_name,document,birthdate,number)]

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                client_message = safe_socket.recv_all(
                    client_socket, _ECHO_SERVER_MESSAGE_SIZE
                )
                if not client_message:
                    logger.info(
                        action,
                        logger.LogResult.success,
                        "messages-amount",
                        message_amount,
                    )
                    return
                
                bets = self._process_message(client_message)
                self.lottery.store_bets(bets)
                message_amount += 1

                response = self._format_response(bets[0])
                safe_socket.send_all(client_socket, response)
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
                    logger.info(action, logger.LogResult.success)
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e

                self._handle_client(client_socket)
