package main

import (
 "bufio"
 "fmt"
 "net"
 "os"
 "regexp"
 "sort"
 "strings"

 router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
 "google.golang.org/protobuf/proto"
)

func main() {
 if err := build(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}

func build() error {
 f, err := os.Open("dist/russia.txt"); if err != nil { return err }; defer f.Close()
 label := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
 domains := map[string]bool{}
 scanner := bufio.NewScanner(f)
 line := 0
 for scanner.Scan() {
  line++
  s := strings.TrimSpace(strings.SplitN(scanner.Text(), "#", 2)[0])
  if s == "" { continue }
  s = strings.ToLower(strings.TrimSuffix(s, "."))
  if len(s) > 253 || !strings.Contains(s, ".") || net.ParseIP(s) != nil { return fmt.Errorf("invalid domain at line %d: %q", line, s) }
  for _, p := range strings.Split(s, ".") { if !label.MatchString(p) { return fmt.Errorf("invalid domain at line %d: %q", line, s) } }
  domains[s] = true
 }
 if err := scanner.Err(); err != nil { return err }
 if len(domains) == 0 { return fmt.Errorf("empty domain list") }
 names := make([]string, 0, len(domains))
 for s := range domains { names = append(names, s) }; sort.Strings(names)
 site := &router.GeoSite{CountryCode: "fptn-blacklist"}
 for _, s := range names { site.Domain = append(site.Domain, &router.Domain{Type: router.Domain_RootDomain, Value: s}) }
 data, err := proto.Marshal(&router.GeoSiteList{Entry: []*router.GeoSite{site}}); if err != nil { return err }
 var checked router.GeoSiteList
 if err := proto.Unmarshal(data, &checked); err != nil { return err }
 if len(checked.Entry) != 1 || checked.Entry[0].CountryCode != "fptn-blacklist" || len(checked.Entry[0].Domain) != len(names) { return fmt.Errorf("verification failed") }
 if err := os.WriteFile("dist/geosite.dat", data, 0644); err != nil { return err }
 if err := os.WriteFile("dist/domains.txt", []byte(strings.Join(names, "\n")+"\n"), 0644); err != nil { return err }
 fmt.Printf("Verified geosite.dat: tag=fptn-blacklist domains=%d bytes=%d\n", len(names), len(data))
 return nil
}
