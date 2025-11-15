import request from './request'

export const sendPlayCommand = (serverID: number, deviceID: number, command: {
  command: string
  session_id: string
  position?: number
}) => {
  return request.post(`/api/playback/${serverID}/${deviceID}/command`, command)
}

export const getActiveSessions = (serverID: number) => {
  return request.get('/api/playback/sessions', {
    params: { server_id: serverID }
  })
}

export const getPlaybackHistory = (params: {
  limit?: number
  offset?: number
}) => {
  return request.get('/api/playback/history', { params })
}
