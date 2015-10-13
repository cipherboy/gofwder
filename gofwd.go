package main

import (
  "io"
  "os"
  "net"
  "log"
  "sync"
)

var wg sync.WaitGroup

func main() {
  if len(os.Args) < 4 || (len(os.Args) - 1) % 3 != 0 {
    help()
    return
  }
  
  for i := 1; i < len(os.Args)-2; i+=3 {
    wg.Add(1)
    listen(os.Args[i], os.Args[i+1], os.Args[i+2])
  }
  wg.Wait()
}

func listen(local_host string, listen_type string, remote_host string) {
  go func() {
    if listen_type == "tcp" || listen_type == "tcp4" || listen_type == "tcp6" {
      local_addr, err := net.ResolveTCPAddr(listen_type, local_host)
      if local_addr == nil {
        log.Println("net.ResolveTCPAddr failed: ", err)
        log.Println("Dropping ", local_host)
        wg.Done()
        return
      }

      remote_addr, err := net.ResolveTCPAddr(listen_type, remote_host)
      if remote_addr == nil {
        log.Println("net.ResolveTCPAddr failed: ", err)
        log.Println("Dropping ", remote_host)
        wg.Done()
        return
      }
      remote_addr = nil

      error_count := 0

      for error_count < 20 {
        local_listener, err := net.ListenTCP(listen_type, local_addr)
        if local_listener == nil {
          log.Println("net.ListenTCP failed: ", err)
          error_count += 1
        }

        log.Println("Listening on " + local_host + " [" + listen_type + "], bound to " + remote_host)

        for {
          active_conn, err := local_listener.Accept()
          if active_conn == nil {
            log.Println("net.TCPListener.Accept() failed: ", err)
            error_count += 1
            break
          }

          go forward_tcp(active_conn, listen_type, remote_host)
        }
      }
    } else if listen_type == "udp" || listen_type == "udp4" || listen_type == "udp6" {
      local_addr, err := net.ResolveUDPAddr(listen_type, local_host)
      if local_addr == nil {
        log.Println("net.ResolveUDPAddr failed: ", err)
        log.Println("Dropping ", local_host)
        wg.Done()
        return
      }

      remote_addr, err := net.ResolveUDPAddr(listen_type, remote_host)
      if remote_addr == nil {
        log.Println("net.ResolveUDPAddr failed: ", err)
        log.Println("Dropping ", remote_host)
        wg.Done()
        return
      }

      error_count := 0

      for error_count < 20 {
        local_listener, err := net.ListenUDP(listen_type, local_addr)
        if local_listener == nil {
          log.Println("net.ListenUDP failed: ", err)
          error_count += 1
        }

        log.Println("Listening on " + local_host + " [" + listen_type + "], bound to " + remote_host)

        for {

          go forward_udp(active_conn, listen_type, remote_host)
        }
      }
    } else {
      log.Println("Unknown listen type: " + listen_type)
    }
    wg.Done()
  }()
}

func forward_tcp(local_conn net.Conn, listen_type string, remote_host string) {
  remote_conn, err := net.Dial(listen_type, remote_host)
  if remote_conn == nil {
    log.Println("net.Dial failed: ", err)
    log.Println("Dropping connection to ", remote_host)
    return
  }
  
  go func() {
    wg.Add(1)
    defer local_conn.Close()
    defer remote_conn.Close()
    io.Copy(local_conn, remote_conn)
    wg.Done()
  }()
  
  go func() {
    wg.Add(1)
    defer local_conn.Close()
    defer remote_conn.Close()
    io.Copy(remote_conn, local_conn)
    wg.Done()
  }()
}


func help() {
  log.Println("Usage: gofwder <local:port> <type> <remote:port> [<local:port> <type> <remote:port> ...]\n")
  log.Println("type: tcp,tcp4,tcp6")
}