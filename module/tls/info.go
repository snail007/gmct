package tlstool

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type certInfo map[string]string

func parseCert(cert *x509.Certificate) (certInfo, error) {
	c := make(certInfo, 0)

	// signature
	dst := make([]byte, hex.EncodedLen(len(cert.Signature)))
	hex.Encode(dst, cert.Signature)
	c["Signature"] = strings.ToUpper(string(dst))
	c["SignatureAlgorithm"] = cert.SignatureAlgorithm.String()
	c["PublicKeyAlgorithm"] = cert.PublicKeyAlgorithm.String()
	c["Version"] = strconv.FormatInt(int64(cert.Version), 10)
	c["SerialNumber"] = cert.SerialNumber.String()
	c["Issuer"] = cert.Issuer.String()
	c["KeyUsage"] = keyUsage(cert.KeyUsage)
	c["ExtKeyUsage"] = extKeyUsage(cert.ExtKeyUsage)
	c["Subject"] = cert.Subject.String()
	c["NotBefore"] = cert.NotBefore.String()
	c["NotAfter"] = cert.NotAfter.String()
	c["IsCA"] = strconv.FormatBool(cert.IsCA)

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		pub := cert.PublicKey.(*rsa.PublicKey)
		c["PublicKey"] = strings.ToUpper(pub.N.String())
	case *ecdsa.PublicKey:
		pub := cert.PublicKey.(*ecdsa.PublicKey)
		c["PublicKey"] = strings.ToUpper(pub.X.String())
	default:
		c["PublicKey"] = "it's something else"
	}

	// SubjectKeyId
	dst = make([]byte, hex.EncodedLen(len(cert.SubjectKeyId)))
	hex.Encode(dst, cert.SubjectKeyId)
	c["SubjectKeyId"] = strings.ToUpper(string(dst))

	// AuthorityKeyId
	dst = make([]byte, hex.EncodedLen(len(cert.AuthorityKeyId)))
	hex.Encode(dst, cert.AuthorityKeyId)
	c["AuthorityKeyId"] = strings.ToUpper(string(dst))

	if len(cert.DNSNames) > 0 {
		c["DNSNames"] = strings.Join(cert.DNSNames, ",")
	}

	if len(cert.EmailAddresses) > 0 {
		c["EmailAddresses"] = strings.Join(cert.EmailAddresses, ",")
	}

	// IPAddresses
	ips := make([]string, 0)
	for _, ip := range cert.IPAddresses {
		ips = append(ips, ip.String())
	}
	if len(ips) > 0 {
		c["IPAddresses"] = strings.Join(ips, ",")
	}

	// URIs
	uris := make([]string, 0)
	for _, uri := range cert.URIs {
		uris = append(uris, uri.String())
	}
	if len(uris) > 0 {
		c["URIs"] = strings.Join(uris, ",")
	}

	if len(cert.CRLDistributionPoints) > 0 {
		c["CRLDistributionPoints"] = strings.Join(cert.CRLDistributionPoints, ",")
	}

	return c, nil
}

func (ci certInfo) String() string {
	certInfoStr := make([]string, 0)
	order := []string{"Subject", "DNSNames", "NotBefore", "NotAfter",
		"IsCA","Issuer","SerialNumber","Version", "SubjectKeyId", "AuthorityKeyId",
		"KeyUsage", "ExtKeyUsage","EmailAddresses", "IPAddresses", "URIs","CRLDistributionPoints",
		"PublicKeyAlgorithm", "PublicKey", "SignatureAlgorithm","Signature"}
	for _, o := range order {
		val, ok := ci[o]
		if ok {
			certInfoStr = append(certInfoStr, fmt.Sprintf("    %s: %s\n", o, val))
		}
	}
	return strings.Join(certInfoStr, "")
}

func parseCerts(certs []*x509.Certificate) (string, error) {
	o := make([]string, 0)
	for _, c := range certs {
		co, err := parseCert(c)
		if err != nil {
			return "", err
		}
		o = append(o, co.String())
	}
	return strings.Join(o, "\n\n"), nil
}

type tlsInfo map[string]string

func getTLSInfo(cs statusWrapper) (tlsInfo, error) {
	info := make(map[string]string, 0)
	info["Version"] = tlsVersion(cs.Version)
	info["HandshakeComplete"] = strconv.FormatBool(cs.HandshakeComplete)
	info["CipherSuite"] = cipherSuite(cs.CipherSuite)
	info["RequestedServerName"] = cs.ServerName
	info["ClientCertificateRequired"] = fmt.Sprintf("%v", cs.requireClientCert)
	// certs
	peerCerts, err := parseCerts(cs.PeerCertificates)
	if err != nil {
		return info, err
	}
	if len(peerCerts) > 0 {
		info["ResponseCertificates"] = fmt.Sprintf("\n%s", peerCerts)
	}
	return info, nil
}

func (ti tlsInfo) String() string {
	tlsInfoStr := make([]string, 0)
	order := []string{"RequestedServerName", "Version", "CipherSuite",
		"HandshakeComplete", "ClientCertificateRequired", "ResponseCertificates"}
	for _, o := range order {
		val, ok := ti[o]
		if ok {
			tlsInfoStr = append(tlsInfoStr, fmt.Sprintf("%s: %s\n", o, val))
		}
	}
	return strings.Join(tlsInfoStr, "")
}

func tlsVersion(c uint16) string {
	versions := map[uint16]string{tls.VersionTLS10: "TLS 1.0",
		tls.VersionTLS11: "TLS 1.1", tls.VersionTLS12: "TLS 1.2",
		tls.VersionTLS13: "TLS 1.3", tls.VersionSSL30: "SSL 3.0"}
	version := fmt.Sprintf("%d", c)
	if v, ok := versions[c]; ok {
		version = v
	}
	return version
}

func cipherSuite(c uint16) string {
	suites := map[uint16]string{
		tls.TLS_RSA_WITH_RC4_128_SHA:                      "TLS_RSA_WITH_RC4_128_SHA",
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:                 "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:                  "TLS_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:                  "TLS_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256:               "TLS_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256:               "TLS_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:               "TLS_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:              "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:          "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:          "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:                "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:            "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:            "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:       "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:         "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:         "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:       "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:         "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:       "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:   "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256: "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",

		// TLS 1.3 cipher suites.
		tls.TLS_AES_128_GCM_SHA256:       "TLS_AES_128_GCM_SHA256",
		tls.TLS_AES_256_GCM_SHA384:       "TLS_AES_256_GCM_SHA384",
		tls.TLS_CHACHA20_POLY1305_SHA256: "TLS_CHACHA20_POLY1305_SHA256",

		// tls.TLS_FALLBACK_SCSV isn't a standard cipher suite but an indicator
		// that the client is doing version fallback. See RFC 7507.
		tls.TLS_FALLBACK_SCSV: "TLS_FALLBACK_SCSV",
	}
	if v, ok := suites[c]; ok {
		return v
	}
	return fmt.Sprintf("%d", c)
}

func keyUsage(c x509.KeyUsage) string {
	ku := map[x509.KeyUsage]string{
		x509.KeyUsageDigitalSignature:  "KeyUsageDigitalSignature",
		x509.KeyUsageContentCommitment: "KeyUsageContentCommitment",
		x509.KeyUsageKeyEncipherment:   "KeyUsageKeyEncipherment",
		x509.KeyUsageDataEncipherment:  "KeyUsageDataEncipherment",
		x509.KeyUsageKeyAgreement:      "KeyUsageKeyAgreement",
		x509.KeyUsageCertSign:          "KeyUsageCertSign",
		x509.KeyUsageCRLSign:           "KeyUsageCRLSign",
		x509.KeyUsageEncipherOnly:      "KeyUsageEncipherOnly",
		x509.KeyUsageDecipherOnly:      "KeyUsageDecipherOnly",
	}
	u := []string{}
	for k, v := range ku {
		if c&k != 0 {
			u = append(u, v)
		}
	}
	if len(u) > 0 {
		return strings.Join(u, ", ")
	}
	return fmt.Sprintf("%d", c)
}
func extKeyUsage(cs []x509.ExtKeyUsage) string {
	ku := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageAny:                            "ExtKeyUsageAny",
		x509.ExtKeyUsageServerAuth:                     "ExtKeyUsageServerAuth",
		x509.ExtKeyUsageClientAuth:                     "ExtKeyUsageClientAuth",
		x509.ExtKeyUsageCodeSigning:                    "ExtKeyUsageCodeSigning",
		x509.ExtKeyUsageEmailProtection:                "ExtKeyUsageEmailProtection",
		x509.ExtKeyUsageIPSECEndSystem:                 "ExtKeyUsageIPSECEndSystem",
		x509.ExtKeyUsageIPSECTunnel:                    "ExtKeyUsageIPSECTunnel",
		x509.ExtKeyUsageIPSECUser:                      "ExtKeyUsageIPSECUser",
		x509.ExtKeyUsageTimeStamping:                   "ExtKeyUsageTimeStamping",
		x509.ExtKeyUsageOCSPSigning:                    "ExtKeyUsageOCSPSigning",
		x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "ExtKeyUsageMicrosoftServerGatedCrypto",
		x509.ExtKeyUsageNetscapeServerGatedCrypto:      "ExtKeyUsageNetscapeServerGatedCrypto",
		x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "ExtKeyUsageMicrosoftCommercialCodeSigning",
		x509.ExtKeyUsageMicrosoftKernelCodeSigning:     "ExtKeyUsageMicrosoftKernelCodeSigning",
	}
	u := []string{}
	for _, c := range cs {
		for k, v := range ku {
			if c&k != 0 {
				u = append(u, v)
			}
		}
	}
	if len(u) > 0 {
		return strings.Join(u, ", ")
	}
	defaultU := ""
	for _, v := range cs {
		defaultU += " ," + fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("%s", defaultU)
}
