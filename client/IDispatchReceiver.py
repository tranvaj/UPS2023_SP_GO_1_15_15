from typing import List


class IDispatchReceiver:
    def receive_from_dispatcher(self, opcode, reply_status, reply_msg: List[str]):
        pass
