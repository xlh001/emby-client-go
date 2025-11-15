// WebSocket客户端服务

export interface WebSocketMessage {
  type: string
  server_id?: string
  data: any
  timestamp: string
}

export type MessageHandler = (message: WebSocketMessage) => void

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string = ''
  private reconnectTimer: number | null = null
  private reconnectAttempts: number = 0
  private maxReconnectAttempts: number = 5
  private reconnectDelay: number = 3000
  private messageHandlers: Set<MessageHandler> = new Set()
  private isManualClose: boolean = false

  /**
   * 连接WebSocket
   */
  connect(token: string, serverId?: string) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      console.log('WebSocket已连接')
      return
    }

    this.isManualClose = false

    // 构建WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    let wsUrl = `${protocol}//${host}/ws?token=${token}`

    if (serverId) {
      wsUrl += `&server_id=${serverId}`
    }

    this.url = wsUrl

    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        console.log('WebSocket连接已建立')
        this.reconnectAttempts = 0
        this.clearReconnectTimer()
      }

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          this.handleMessage(message)
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
      }

      this.ws.onclose = () => {
        console.log('WebSocket连接已关闭')
        this.ws = null

        // 如果不是手动关闭，尝试重连
        if (!this.isManualClose) {
          this.scheduleReconnect(token, serverId)
        }
      }
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      this.scheduleReconnect(token, serverId)
    }
  }

  /**
   * 断开连接
   */
  disconnect() {
    this.isManualClose = true
    this.clearReconnectTimer()

    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * 发送消息
   */
  send(message: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  /**
   * 添加消息处理器
   */
  onMessage(handler: MessageHandler) {
    this.messageHandlers.add(handler)

    // 返回取消订阅函数
    return () => {
      this.messageHandlers.delete(handler)
    }
  }

  /**
   * 处理接收到的消息
   */
  private handleMessage(message: WebSocketMessage) {
    this.messageHandlers.forEach(handler => {
      try {
        handler(message)
      } catch (error) {
        console.error('消息处理器执行失败:', error)
      }
    })
  }

  /**
   * 安排重连
   */
  private scheduleReconnect(token: string, serverId?: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('WebSocket重连次数已达上限')
      return
    }

    this.clearReconnectTimer()

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts)
    console.log(`将在 ${delay}ms 后尝试重连 (第 ${this.reconnectAttempts + 1} 次)`)

    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectAttempts++
      this.connect(token, serverId)
    }, delay)
  }

  /**
   * 清除重连定时器
   */
  private clearReconnectTimer() {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  /**
   * 获取连接状态
   */
  getState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED
  }

  /**
   * 是否已连接
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }
}

// 导出单例
export const wsClient = new WebSocketClient()

export default wsClient
