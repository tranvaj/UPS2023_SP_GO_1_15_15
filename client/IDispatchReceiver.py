from typing import List


class IDispatchReceiver:   
    def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
        """
        Receives a message from the dispatcher.

        Args:
            opcode (int): The opcode of the message.
            reply_status (int): The status of the reply.
            reply_msg (List[str]): The message received.

        Returns:
            None
        """
        pass
