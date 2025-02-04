// Copyright © by Jeff Foley 2017-2023. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/fatih/color"
	amassnet "github.com/insomn14/amass/net"
	"github.com/insomn14/amass/requests"
)

// Banner is the ASCII art logo used within help output.
const Banner = `        .+++:.            :                             .+++.
      +W@@@@@@8        &+W@#               o8W8:      +W@@@@@@#.   oW@@@W#+
     &@#+   .o@##.    .@@@o@W.o@@o       :@@#&W8o    .@#:  .:oW+  .@#+++&#&
    +@&        &@&     #@8 +@W@&8@+     :@W.   +@8   +@:          .@8
    8@          @@     8@o  8@8  WW    .@W      W@+  .@W.          o@#:
    WW          &@o    &@:  o@+  o@+   #@.      8@o   +W@#+.        +W@8:
    #@          :@W    &@+  &@+   @8  :@o       o@o     oW@@W+        oW@8
    o@+          @@&   &@+  &@+   #@  &@.      .W@W       .+#@&         o@W.
     WW         +@W@8. &@+  :&    o@+ #@      :@W&@&         &@:  ..     :@o
     :@W:      o@# +Wo &@+        :W: +@W&o++o@W. &@&  8@#o+&@W.  #@:    o@+
      :W@@WWWW@@8       +              :&W@@@@&    &W  .o#@@W&.   :W@WWW@@&
        +o&&&&+.                                                    +oooo.`

const (
	// Version is used to display the current version of Amass.
	Version = "v4.2.2"

	// Author is used to display the Amass Project Team.
	Author = "OWASP Amass Project - @owaspamass"

	// Description is the slogan for the Amass Project.
	Description = "In-depth Attack Surface Mapping and Asset Discovery"
)

var (
	// Colors used to ease the reading of program output
	g      = color.New(color.FgHiGreen)
	b      = color.New(color.FgHiBlue)
	y      = color.New(color.FgHiYellow)
	r      = color.New(color.FgHiRed)
	yellow = color.New(color.FgHiYellow).SprintFunc()
	green  = color.New(color.FgHiGreen).SprintFunc()
	blue   = color.New(color.FgHiBlue).SprintFunc()
)

// ASNSummaryData stores information related to discovered ASs and netblocks.
type ASNSummaryData struct {
	Name      string
	Netblocks map[string]int
}

func PrintEnumerationSummary(total int, records []string, target string) {
	// Maps to hold the summarized data
	asns := make(map[string]map[string]interface{}) // ASN -> (organization, netblocks, FQDNs)
	fqdns := make(map[string]string)                // FQDN -> IP

	// Parse the records
	for _, record := range records {
		parts := strings.Split(record, " --> ")
		if len(parts) < 3 {
			continue // Skip malformed records
		}

		left := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[2])

		// Check if the record is an ASN
		if strings.HasSuffix(value, " (Netblock)") {
			// If it's a netblock, associate it with the ASN
			ntblocks := strings.TrimSuffix(value, " (Netblock)")
			for asnID := range asns {
				asns[asnID]["netblocks"] = append(asns[asnID]["netblocks"].([]string), ntblocks)
			}
		} else if strings.HasSuffix(left, "(ASN)") {
			asnID := left[:len(left)-len(" (ASN)")]
			asnDetails := strings.Split(value, " ")
			if len(asnDetails) >= 2 && strings.HasSuffix(value, "(RIROrganization)") {
				asns[asnID] = map[string]interface{}{
					"organization": strings.TrimSuffix(value, " (RIROrganization)"),
					"netblocks":    []string{},
					"fqdns":        []string{},
				}
			}
		} else if strings.HasSuffix(left, "(FQDN)") {
			// If it's a FQDN or IP address, store it
			if strings.HasSuffix(left, "(FQDN)") && strings.HasSuffix(value, "(IPAddress)") {
				fqdns[left] = value
				// Associate FQDN with the ASN
				for asnID := range asns {
					asns[asnID]["fqdns"] = append(asns[asnID]["fqdns"].([]string), left)
				}
			} else {
				fqdns[left] = value
			} 
		} 
	}

	// pad := func(num int, chr string) {
	// 	for i := 0; i < num; i++ {
	// 		b.Fprint(color.Error, chr)
	// 	}
	// }

	// fmt.Fprintln(color.Error)
	// // Print the header information
	// title := "OWASP Amass "
	// site := "https://github.com/insomn14/amass"
	// b.Fprint(color.Error, title+Version)
	// num := 80 - (len(title) + len(Version) + len(site))
	// pad(num, " ")
	// b.Fprintf(color.Error, "%s\n", site)
	// pad(8, "----------")
	// fmt.Fprintf(color.Error, "\n%s%s", yellow(strconv.Itoa(total)), green(" records discovered"))
	// fmt.Fprintln(color.Error)

	// if len(asns) == 0 {
	// 	return
	// }
	// // Another line gets printed
	// pad(8, "----------")
	// fmt.Fprintln(color.Error)

	// Print the summary
	// for asnID, details := range asns {
	// 	// Print ASN details
	// 	netblocks := strings.Join(details["netblocks"].([]string), ", ")
	// 	org := details["organization"]
	// 	fmt.Fprintf(color.Error, "\n%s%s - %s \n\t %s\t %s  %s\n", blue("ASN: "), yellow(asnID), green(org), yellow(netblocks), yellow(strconv.Itoa(len(fqdns))), blue("Subdomain Name(s)"))
	// 	for fqdn, ip := range fqdns {
	// 		if strings.HasSuffix(ip, "(FQDN)") {
	// 			// Clean FQDN -> FQDN to FQDN -> IPAddress
	// 			tmp_ip := fqdns[ip]
	// 			fmt.Fprintf(color.Error, "\n%s --> %s", green(strings.TrimSuffix(fqdn, " (FQDN)")), yellow(strings.TrimSuffix(tmp_ip, " (IPAddress)")))
	// 		} else {
	// 			fmt.Fprintf(color.Error, "\n%s --> %s", green(strings.TrimSuffix(fqdn, " (FQDN)")), yellow(strings.TrimSuffix(ip, " (IPAddress)")))
	// 		}
	// 	}
	// }
	// PrintASNDetails(asns, fqdns)

	// Generate dynamic filename with current date
	currentDate := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s_%s.txt", target, currentDate)
	if err := SaveASNDetailsToFile(filename, asns, fqdns); err != nil {
		color.Red("\n[!] Error saving file: %v", err)
	} else {
		color.Green("\n[+] Details saved to %s", filename)
	}
}

// PrintASNDetails prints ASN details to the console
func PrintASNDetails(asns map[string]map[string]interface{}, fqdns map[string]string) {
	for asnID, details := range asns {
		// Print ASN details
		org := details["organization"].(string)
		netblocks := strings.Join(details["netblocks"].([]string), ", ")
		fmt.Fprintf(color.Error, "\n%s%s - %s\n\t%s%s\t%s%s\n",
			color.BlueString("ASN: "), color.YellowString(asnID), color.GreenString(org),
			color.YellowString(netblocks), color.YellowString(strconv.Itoa(len(fqdns))), color.BlueString(" Subdomain Name(s)"))

		// Print FQDNs and associated IPs
		for fqdn, ip := range fqdns {
			if strings.HasSuffix(ip, "(FQDN)") {
				tmpIP := fqdns[ip]
				fmt.Fprintf(color.Error, "\n%s --> %s",
					color.GreenString(strings.TrimSuffix(fqdn, " (FQDN)")),
					color.YellowString(strings.TrimSuffix(tmpIP, " (IPAddress)")))
			} else {
				fmt.Fprintf(color.Error, "\n%s --> %s",
					color.GreenString(strings.TrimSuffix(fqdn, " (FQDN)")),
					color.YellowString(strings.TrimSuffix(ip, " (IPAddress)")))
			}
		}
	}
	fmt.Fprintln(color.Error)
}

func SaveASNDetailsToFile(filename string, asns map[string]map[string]interface{}, fqdns map[string]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	for asnID, details := range asns {
		// Write ASN details
		org := details["organization"].(string)
		netblocks := strings.Join(details["netblocks"].([]string), ", ")
		_, err := file.WriteString(fmt.Sprintf("ASN: %s - %s\n\tNetblocks: %s\n\tSubdomains: %d\n", asnID, org, netblocks, len(fqdns)))
		if err != nil {
			return fmt.Errorf("failed to write ASN details: %v", err)
		}

		// Write FQDNs and associated IPs
		for fqdn, ip := range fqdns {
			var line string
			if strings.HasSuffix(ip, "(FQDN)") {
				tmpIP := fqdns[ip]
				line = fmt.Sprintf("%s: %s\n", strings.TrimSuffix(fqdn, " (FQDN)"), strings.TrimSuffix(tmpIP, " (IPAddress)"))
			} else {
				line = fmt.Sprintf("%s: %s\n", strings.TrimSuffix(fqdn, " (FQDN)"), strings.TrimSuffix(ip, " (IPAddress)"))
			}
			if _, err := file.WriteString(line); err != nil {
				return fmt.Errorf("failed to write FQDN details: %v", err)
			}
		}
	}
	return nil
}

// UpdateSummaryData updates the summary maps using the provided requests.Output data.
func UpdateSummaryData(output *requests.Output, asns map[int]*ASNSummaryData) {
	for _, addr := range output.Addresses {
		if addr.CIDRStr == "" {
			continue
		}

		data, found := asns[addr.ASN]
		if !found {
			asns[addr.ASN] = &ASNSummaryData{
				Name:      addr.Description,
				Netblocks: make(map[string]int),
			}
			data = asns[addr.ASN]
		}
		// Increment how many IPs were in this netblock
		data.Netblocks[addr.CIDRStr]++
	}
}

// PrintEnumerationSummary outputs the summary information utilized by the command-line tools.
// func PrintEnumerationSummary(total int, asns map[int]*ASNSummaryData, demo bool) {
// 	FprintEnumerationSummary(color.Error, total, asns, demo)
// }


// FprintEnumerationSummary outputs the summary information utilized by the command-line tools.
func FprintEnumerationSummary(out io.Writer, total int, asns map[int]*ASNSummaryData, demo bool) {
	pad := func(num int, chr string) {
		for i := 0; i < num; i++ {
			b.Fprint(out, chr)
		}
	}

	fmt.Fprintln(out)
	// Print the header information
	title := "OWASP Amass "
	site := "https://github.com/insomn14/amass"
	b.Fprint(out, title+Version)
	num := 80 - (len(title) + len(Version) + len(site))
	pad(num, " ")
	b.Fprintf(out, "%s\n", site)
	pad(8, "----------")
	fmt.Fprintf(out, "\n%s%s", yellow(strconv.Itoa(total)), green(" names discovered"))
	fmt.Fprintln(out)

	if len(asns) == 0 {
		return
	}
	// Another line gets printed
	pad(8, "----------")
	fmt.Fprintln(out)
	// Print the ASN and netblock information
	for asn, data := range asns {
		asnstr := strconv.Itoa(asn)
		datastr := data.Name

		if demo && asn > 0 {
			asnstr = censorString(asnstr, 0, len(asnstr))
			datastr = censorString(datastr, 0, len(datastr))
		}
		fmt.Fprintf(out, "%s%s %s %s\n", blue("ASN: "), yellow(asnstr), green("-"), green(datastr))

		for cidr, ips := range data.Netblocks {
			countstr := strconv.Itoa(ips)
			cidrstr := cidr

			if demo {
				cidrstr = censorNetBlock(cidrstr)
			}

			countstr = fmt.Sprintf("\t%-4s", countstr)
			cidrstr = fmt.Sprintf("\t%-18s", cidrstr)
			fmt.Fprintf(out, "%s%s %s\n", yellow(cidrstr), yellow(countstr), blue("Subdomain Name(s)"))
		}
	}
}


// PrintBanner outputs the Amass banner to stderr.
func PrintBanner() {
	FprintBanner(color.Error)
}

// FprintBanner outputs the Amass banner the same for all tools.
func FprintBanner(out io.Writer) {
	rightmost := 76

	pad := func(num int) {
		for i := 0; i < num; i++ {
			fmt.Fprint(out, " ")
		}
	}

	_, _ = r.Fprintf(out, "\n%s\n\n", Banner)
	pad(rightmost - len(Version))
	_, _ = y.Fprintln(out, Version)
	pad(rightmost - len(Author))
	_, _ = y.Fprintln(out, Author)
	pad(rightmost - len(Description))
	_, _ = y.Fprintf(out, "%s\n\n\n", Description)
}

func censorDomain(input string) string {
	return censorString(input, strings.Index(input, "."), len(input))
}

func censorIP(input string) string {
	return censorString(input, 0, strings.LastIndex(input, "."))
}

func censorNetBlock(input string) string {
	return censorString(input, 0, strings.Index(input, "/"))
}

func censorString(input string, start, end int) string {
	runes := []rune(input)
	for i := start; i < end; i++ {
		if runes[i] == '.' ||
			runes[i] == '/' ||
			runes[i] == '-' ||
			runes[i] == ' ' {
			continue
		}
		runes[i] = 'x'
	}
	return string(runes)
}

// OutputLineParts returns the parts of a line to be printed for a requests.Output.
func OutputLineParts(out *requests.Output, addrs, demo bool) (name, ips string) {
	if addrs {
		for i, a := range out.Addresses {
			if i != 0 {
				ips += ","
			}
			if demo {
				ips += censorIP(a.Address.String())
			} else {
				ips += a.Address.String()
			}
		}
	}
	name = out.Name
	if demo {
		name = censorDomain(name)
	}
	return
}

func OutputLinePartsOld(out *requests.Output, src, addrs, demo bool) (source, name, ips string) {
	if src {
		source = fmt.Sprintf("%-18s", "["+out.Sources[0]+"] ")
	}
	if addrs {
		for i, a := range out.Addresses {
			if i != 0 {
				ips += ","
			}
			if demo {
				ips += censorIP(a.Address.String())
			} else {
				ips += a.Address.String()
			}
		}
		if ips == "" {
			ips = "N/A"
		}
	}
	name = out.Name
	if demo {
		name = censorDomain(name)
	}
	return
}


// DesiredAddrTypes removes undesired address types from the AddressInfo slice.
func DesiredAddrTypes(addrs []requests.AddressInfo, ipv4, ipv6 bool) []requests.AddressInfo {
	var kept []requests.AddressInfo

	for _, addr := range addrs {
		if ipv4 && amassnet.IsIPv4(addr.Address) {
			kept = append(kept, addr)
		} else if ipv6 && amassnet.IsIPv6(addr.Address) {
			kept = append(kept, addr)
		}
	}

	return kept
}

// InterfaceInfo returns network interface information specific to the current host.
func InterfaceInfo() string {
	var output string

	if ifaces, err := net.Interfaces(); err == nil {
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				continue
			}
			output += fmt.Sprintf("%s%s%s\n", blue(i.Name+": "), green("flags="), yellow("<"+strings.ToUpper(i.Flags.String()+">")))
			if i.HardwareAddr.String() != "" {
				output += fmt.Sprintf("\t%s%s\n", green("ether: "), yellow(i.HardwareAddr.String()))
			}
			for _, addr := range addrs {
				inet := "inet"
				if a, ok := addr.(*net.IPNet); ok && amassnet.IsIPv6(a.IP) {
					inet += "6"
				}
				inet += ": "
				output += fmt.Sprintf("\t%s%s\n", green(inet), yellow(addr.String()))
			}
		}
	}

	return output
}
