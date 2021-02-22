package main

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"ntc.org/mclib/common"
)

type AWSRoute53 struct {
	profile common.AwsConfig
	config  aws.Config
	session *session.Session
	dnsMgr  *route53.Route53
}

func NewAWSRoute53(config common.AwsConfig) *AWSRoute53 {
	ac := &aws.Config{
		Region: aws.String(config.RegionId),
		Credentials: credentials.NewStaticCredentials(config.AccessId, config.SecretKey, config.SessionToken),
	}
	sess := session.Must(session.NewSession(ac))
	return &AWSRoute53{
		profile: config,
		config:  *ac,
		session: sess,
		dnsMgr:  route53.New(sess),
	}
}

type dnsRequest struct {
	Name   string
	Target string
	TTL    int64
	Weight int64
	ZoneId string
}
func (r *AWSRoute53) GetDomainRecords(request dnsRequest) ([]*route53.ResourceRecordSet, error){
	return r.listDomainRecords(request)
}
func (r *AWSRoute53) listDomainRecords(request dnsRequest) ([]*route53.ResourceRecordSet, error){
	zoneId := request.ZoneId
	if zoneId == ""{
		zoneId = r.profile.DNSZoneId
	}
	listParams := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneId), // Required
	}
	respList, err := r.dnsMgr.ListResourceRecordSets(listParams)
	var list []*route53.ResourceRecordSet
	if request.Name!=""{
		n := strings.Trim(request.Name, ".;:") + "."
		for _, r := range respList.ResourceRecordSets{
			if r.Name==nil || *r.Name != n{
				continue
			}
			list = append(list, r)
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to List Domain Records")
	}
	return list, err
}
func (r *AWSRoute53) LookupDomainIP(request dnsRequest) (string, error){
	resp, err := r.LookupDomainIPs(request)
	if err != nil{
		return "", err
	}
	for _, r := range resp{
		return r, nil
	}
	return "", nil
}
func (r *AWSRoute53) LookupDomainIPs(request dnsRequest) ([]string, error){
	resp, err := r.listDomainRecords(request)
	if err!=nil{
		return nil, err
	}
	ips := []string{}
	for _, r := range resp{
		if len(r.ResourceRecords) < 1{
			continue
		}
		for _, ra := range r.ResourceRecords{
			if ra.Value == nil || *ra.Value==""{
				continue
			}
			if !common.REIpAddress.MatchString(*ra.Value){
				continue
			}
			ips = append(ips, *ra.Value)
		}
	}
	return ips, nil
}
func (r *AWSRoute53) UpsertDomainRecord(request dnsRequest) (*route53.ChangeResourceRecordSetsOutput, error){
	_, err := r.DeleteDNSRecord(request)
	if err!=nil{
		return nil, err
	}
	return r.upsertResourceRecord(request)
}
func (r *AWSRoute53) upsertResourceRecord(request dnsRequest) (*route53.ChangeResourceRecordSetsOutput, error){
	zoneId := request.ZoneId
	if zoneId == ""{
		zoneId = r.profile.DNSZoneId
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("UPSERT"), // Required
					ResourceRecordSet: &route53.ResourceRecordSet{ // Required
						Name: aws.String(request.Name),    // Required
						Type: aws.String("A"), // Required
						ResourceRecords: []*route53.ResourceRecord{
							{ // Required
								Value: aws.String(request.Target), // Required
							},
						},
						TTL:           aws.Int64(request.TTL),
						Weight:        aws.Int64(255),
						SetIdentifier: aws.String("MCL:" + request.Name),
					},
				},
			},
			//Comment: aws.String("Sample update."),
		},
		HostedZoneId: aws.String(zoneId), // Required
	}
	log.Info().Msgf("Created DNS Record: %s -> %s", request.Name, request.Target)
	resp, err := r.dnsMgr.ChangeResourceRecordSets(params)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to Upsert Domain Cname Record")
	}
	return resp, nil
}
func (r *AWSRoute53) DeleteDNSRecord(request dnsRequest) ([]*route53.ChangeResourceRecordSetsOutput, error){
	zoneId := request.ZoneId
	if zoneId == ""{
		zoneId = r.profile.DNSZoneId
	}
	if request.Name==""{
		return nil, errors.Errorf("You may delete dns records by domain name")
	}
	list, err := r.listDomainRecords(request)
	if err != nil{
		return nil, err
	}
	var resp []*route53.ChangeResourceRecordSetsOutput
	for _, d := range list{
		if d == nil{
			continue
		}
		o, err := r.deleteResourceRecordSet(zoneId, *d)
		if err != nil{
			return resp, err
		}
		log.Info().Msgf("Deleted DNS Record: %s", request.Name)
		resp = append(resp, o)
	}
	return resp, nil
}
func (r *AWSRoute53) deleteResourceRecordSet(zoneId string, request route53.ResourceRecordSet) (*route53.ChangeResourceRecordSetsOutput, error){
	if zoneId == ""{
		zoneId = r.profile.DNSZoneId
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{ // Required
			Changes: []*route53.Change{ // Required
				{ // Required
					Action: aws.String("DELETE"), // Required
					ResourceRecordSet: &request,
				},
			},
			//Comment: aws.String("Sample update."),
		},
		HostedZoneId: aws.String(zoneId), // Required
	}
	resp, err := r.dnsMgr.ChangeResourceRecordSets(params)

	if err != nil {
		return nil, errors.Wrapf(err, "Unable to Upsert Domain Cname Record")
	}
	return resp, nil
}
