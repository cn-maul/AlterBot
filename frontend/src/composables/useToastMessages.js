import { ref, onUnmounted } from 'vue'

export function useToastMessages() {
  const successMsg = ref('')
  const pageErrorMsg = ref('')
  let msgTimer = null

  onUnmounted(() => {
    clearTimeout(msgTimer)
  })

  function showSuccess(msg) {
    successMsg.value = msg
    clearTimeout(msgTimer)
    msgTimer = setTimeout(() => { successMsg.value = '' }, 3000)
  }

  function showError(msg) {
    pageErrorMsg.value = msg
    clearTimeout(msgTimer)
    msgTimer = setTimeout(() => { pageErrorMsg.value = '' }, 5000)
  }

  return { successMsg, pageErrorMsg, showSuccess, showError }
}