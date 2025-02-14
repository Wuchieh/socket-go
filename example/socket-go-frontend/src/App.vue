<template>
  <div class="container">
    <dialog ref="dialogRef">
      <div>
        <form @submit.prevent="onSubmitName">
          <s-input label="請輸入暱稱" v-model="name" class="mb-2"/>
          <s-btn>送出</s-btn>
        </form>
      </div>
    </dialog>

    <div v-if="!socketIsConnected">
      <h1>連線中...</h1>
    </div>

    <div>
      <h1 v-if="name">你好 {{ name }}</h1>
      <div>{{ atRoom ? `目前所在房間: ${atRoom}` : '尚未加入房間' }}</div>
      <div>
        <button @click="join('room_1')">房間1</button>
        <button @click="join('room_2')">房間2</button>
      </div>
      <div>
        <input type="text" v-model="msg">
        <button @click="sendMsg">發送</button>
      </div>
      <div>
        <div v-for="v in messages">
          {{ v }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {nextTick, onMounted, ref} from "vue";
import SInput from "./components/SInput.vue";
import SBtn from "./components/SBtn.vue";
import {sleep} from "./utils";
import {nameStore, type Socket, useSocket} from "./composables";

const socket = ref<Socket | null>(null)
const atRoom = ref("")
const socketIsConnected = ref(false)
const dialogRef = ref<HTMLDialogElement | null>(null)
const name = ref("")
const msg = ref("")
const messages = ref<{
  name: string
  message: string
  time: string
}[]>([])

const sendMsg = () => {
  if (!msg.value) return
  socket.value?.emit("send_massage", msg.value)
  msg.value = ""
}

const join = (room: string) => {
  socket.value?.emit("join_room", room)
}

const onSubmitName = () => {
  const _name = name.value
  if (!_name) return

  socket.value?.emit("set_name", _name)
  nameStore.set(_name)
  hiddenDialog()
}

const showDialog = () => {
  const dialog = dialogRef.value
  if (!dialog) return

  dialog.showModal()
  nextTick(() => {
    dialog.classList.add("show")
  })
}

const hiddenDialog = () => {
  const dialog = dialogRef.value
  if (!dialog) return

  dialog.classList.remove("show")
  sleep(400).then(() => {
    dialog.close()
  })

}

function initName() {
  const _name = nameStore.get()
  if (!_name) {
    showDialog()
    return
  }
  name.value = _name
  socket.value!.emit("set_name", _name)
}

function initSocket() {
  socket.value = useSocket()
  socket.value.on("send_massage", v => {
    messages.value.push(v!)
  })
  socket.value.on("join_room", v => {
    atRoom.value = v!
  })
  return new Promise((resolve) => {
    socket.value!.onConnect(() => resolve(void 0))
  })
}

function init() {
  initSocket().then(() => {
    socketIsConnected.value = true
    initName()
  })
}

onMounted(() => {
  init()
})
</script>

<style scoped>
</style>