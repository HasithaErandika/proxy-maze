

TORCHLABSSRILANKA2026EngineeringChallenge
ProxyMaze'26
TheReal-TimeProxyIntelligenceChallenge
Buildthewatchtowerweshouldhavehadayearago.
## CHALLENGEOVERVIEW
## 250   +20180
CorePointsBonusPointsPassingScore
SubmittedtoengineeringcandidatesacrossSriLanka.
Evaluatedentirelyasablackb ox.YourHTTPAPIisallwewilleversee.
TorchLabs•Colomb o,SriLanka•FromSriLanka,totheworld.

ProxyMaze'26TheTorchLabsChallenge1
## Contents
1WheretheStoryBegins2
2ThePeople3
3TheNightItAllWentWrong4
4TheWhiteb oard6
5WhyYou7
5.1WhatProxyMazeMustDo.............................7
6ScoreSummary8
7ProxyIdentiersandGroundRules9
Chapter01:Pro ofofLife10
Chapter02:TheHeartb eat10
Chapter03:TheMemory10
Chapter04:BuildingthePo ol11
Chapter05:TheWatchtower11
Chapter06:TheDossier12
Chapter07:TheChronicle12
Chapter08:TheGraveyard13
Chapter09:TheAlertArchive13
7.0.1Alertob ject:requiredelds.........................13
Chapter10:TheMessenger14
## 7.0.2alert.firedpayload............................14
## 7.0.3alert.resolvedpayload..........................14
7.0.4Deliveryrequirements............................15
Chapter11:TheIntegrationLayer15
Chapter12:TheControlRo om15
8BehavioralRules17
8.1AlertLifecycle....................................17
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge2
## PARTONE
TheTorchLabsStory
1WheretheStoryBegins
ThereisabuildinginColombo,aconvertedwarehouseonaquietstretchofroadbehindtheold
railwayyard,wheretheairalwayssmel lsfaintlyofstrongtea,solderingux,andambition.The
rst-oorwindowsfaceeast,andonearlymorningsthesunlightspil lsatacrosslongwooden
deskscoveredinlaptops,notebooks,andthekindoftangledethernetcablesthatsay,without
saying,thatthisisaplacewherethingsactual lygetbuilt.
ThisisTorchLabs.
Itstartedwithfourclassmatesandanargumentoverchai.Theargumentwassimple,butithad
beenbrewingforyears:SriLankanengineerswerebril liant,thegroupagreed,butthecountry
hadbeenquietlyexportingthatbril lianceforadecade.Everysenior,everypostgraduate,every
unusual lysharpjunior.Gone.ToSingapore.ToSydney.ToLondon.ToSanFrancisco.The
countrywasproducingworld-classengineersandshippingthemoverseaslikecinnamonandtea.
Thefourofthemdecidedthatiftheyweregoingtopushbackagainstthattrajectory,theywould
havetodoitwithacompany.
ThecompanytheybuiltisTorchLabs.Theproducttheychose,theproblemtheyplantedtheir
agin,istheunglamorous,deeply-plumbing-shapedworldofinternetproxyinfrastructure.
Therearetwoagshipproducts.TherstisResidentialEdge,aeetofgenuineresidential
IPaddressessourcedfromrealdevicesinfteencountries:thekindoftracthatdoesn'ttrip
detectionsystemsbecause,technical ly,itisn'tsuspicioustracatal l.ThesecondisISPCore,
aworkhorselineofdedicatedISP-gradeproxiestunedforhighthroughputandusedheavilyby
marketintel ligenceanddataanalyticsclientsacrossSoutheastAsia,EastAsia,andpartsof
## Europe.
Thecompanyisprotable.Itisalsosmal lenoughthateveryoneknowseveryoneelse'scoee
order.Thereareaboutthirtyengineers.ThereisnoNOCteamintheformalsense;thereisa
Slackchannelcal led#nocthatl lsupatthreeinthemorningwhensomethingunusualhappens
inFrankfurtorManilaorJakarta.Thereisawhiteboardatthebackofthesecond-ooroce
thatisnevererased,becauseeveryoneisafraidoferasingthewrongline.
Thecompanyisalsogrowing.Thousandsofclients.Thousandsofactiveproxyendpoints.
Hundredsofthousandsofrequestsperminuteatpeak.And,untilveryrecently,exactlyone
personwhosejobitwastoknowwhetheranyofthoseendpointswereactual lyworking.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge3
2ThePeople
SachinistheCTO.HegrewupinGal le,studiedcomputerscienceattheUniversityofWayamba,
andspentfouryearsatadistributedsystemsstartupinIrelandbeforehecamehomewithaduel
bagandanidea.Hisengineersdescribehimintwowords,dependingonthemood:methodical
onthedayswhenthesystemiscalm,andterrifyingonthedayswhenitisn't.Herunsthe
engineeringorgwithasoftvoiceandanear-religiousbeliefinpostmortems.Hekeepsasmal l,
rulednotebookinhisbackpocket;theteamhasneverseenwhatiswritteninit,buteveryonehas
seenhimpul litoutduringincidentsandwriteasingleline,veryslow ly,beforesayinganything.
Hirushaistheseniorinfrastructureengineer.HejoinedTorchLabseighteenmonthsago,after
veyearsbuildingsearchinfrastructureforane-commercecompanyinSingapore.Hecame
back,hetoldSachinduringtheinterview,becausemyparentsaregettingolderandbecauseSri
LankaisgoingtobereadyforseriousinfrastructurecompaniesinaboutveyearsandI'drather
beearlythanlate.Sachinhiredhimonthespot.
Hirushaistheengineerwhosephonegetstherstcal lwhensomethingbreaksat11:47atnight.
Hehasawife,aone-year-olddaughternamedSenuli,andahabitofeatinginstantnood lesat
hisdeskbecauseheforgetstoleavefordinner.Heisthehumansafetynetforthousandsofproxy
endpoints.Heisalso,increasingly,exhausted.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge4
3TheNightItAllWentWrong
## Thechronologyoftheincident,aslaterreconstructedforthepostmortem,wentlikethis.
Tuesday,11:47PM,Colombo.Hirushawashalfwaythroughabow lofinstantnood lesathis
kitchentablewhenhisphonelitupwithaSlacknotication.Theclientwasadataanalyticsrm
inSeoulthatrannightlyscrapingjobsagainstSouthKoreane-commercesites.Themessage,
translatedbyhisphone'sauto-translate,readapproximately:Yourproxiesaretimingout.Al l
ofthem.Ourjobstartsinfourhours.Weneedanswersnow.
Hirushaputdownhischopsticks.Heopenedhislaptop.Heopenedthemonitoringdashboard,
whichwasgenerouslynamed.Itwasahalf-nishedPythonscriptwritteneightmonthsearlierby
asummerinternwhohadsincegraduatedandmovedtoSingapore,wherehewasnowreported ly
makingthreetimesHirusha'ssalary.Thescriptwassupposedtoruneveryfteenminutes.
Accordingtoitslog,ithadlastrunat6:02PM,silentforalmostsixhours.Theintern,when
nal lyreachedat2AMSingaporetime,didnotknowwhy.
12:04AM.Hirushabeganpingingproxiesmanual ly.Heopenedaterminal,copiedtheproxy
listoutofaGoogleSheetthatthesalesteamalsoedited,andstartedrunningcurlcommands.
By12:30AM,hehadninety-sixconrmeddeadproxiesoutoftwohundredandtwentyinthe
client'sdedicatedResidentialEdgepool.Forty-threepercent.Hehadnoideawhenthefailurehad
started.Hehadnoideawhethermoreweredyingrightnow,whilehewaslookingattheones
alreadygone.TheGoogleSheettoldhimwhichproxiesexisted;itdidnottel lhimwhetherthey
werealive.
12:45AM.Herealized,withtheslowcertaintythatarrivesonlyatverylatehours,thatan
upstreamISPservingthreeoftheircountrieshadquietlyupdatedtheirroutingtablesearlierthat
evening.Thetableswerenottechnical lywrong.Theywerejustdierent.Dierentenoughthat
approximatelyhalfofhisresidentialpoolcouldnolongerestablishoutboundconnections.
1:30AM.Hecal ledSachin.Sachinansweredonthesecondring.Hedidn'tsoundsurprised.
Hesounded,instead,likeamanwhohadbeenquietlyexpectingthisphonecal lformonths.
2:15AM.Thetwoofthem,workingfromoppositeendsofthecity,swappedroutingforthe
aectedpooltoafal lbackISP.By3:08AM,ninety-oneoftheninety-sixdeadproxieshadrecov-
ered.Theremainingveweregenuinelydeadandwouldhavetoberetiredinthemorning.
3:34AM.Hirushaclosedhislaptop.HisdaughterSenuliwasawake,gurglingquietlyinher
crib.Theskyoutsidethewindowwasthecolourofweaktea.Hehadbeenawakefornineteen
hours.
Wednesday,9:12AM.TheSeoulclientledaformalSLA-creditrequestforthemissed
scrapingwindow.
Wednesday,11:41AM.Asecondclient,inJakarta,ledarelatedrequest.Theirdashboards
hadshownelevatederrorratesfromroughly8PMthepreviousevening,whichtheyhadassumed
wasatransientissue,untiltheyreadabouttheKoreanclient'ssituationinaprivatechannelon
amarket-dataforum.
Wednesday,4:02PM.TheSeoulclientcancel ledtheirsubscription.Theycitednottheoutage
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge5
itself(outageshappen,theywroteintheircancel lationemail)butthefactthatTorchLabshad
clearlynotknowntheoutagewashappening.Wecantoleratefailure,theemailread.We
cannottolerateavendorwhofailssilently.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge6
4TheWhiteb oard
TheteamgatheredonWednesdayeveninginthesmal lconferenceroomnexttothekitchen.
Thereweresixofthem:Sachin,Hirusha,twobackendengineers,theheadofcustomersuccess,
andajuniorwhohadbeenhiredtheweekbeforeandwhohadspenttheentiremeetinglooking
slightlyhorried.
Sachindidnotsaymuch.Helistened.HeletHirushatalk.Heletthecustomersuccesslead
readaloudthecancel lationemail,twice.Thenhestoodup,walkedovertothewhiteboardatthe
backoftheroom,theonenooneevererased,andwrote,incapitallettersacrossthetop:
## PROXYMAZE
Heunderlinedit.Twice.Then,beneathit,hewrotethreelines:
Itmustwatchthepoolcontinuously,withoutbeingasked.
Itmustknowwhenproxiesfail,andwhichonesareabouttofail.
Itmustwakesomebodyupbeforeaclientemailsus.
Hecappedthemarker.Heturnedaround.
Weneededthissixmonthsago,hesaid.Weneededitayearago.HelookedatHirusha,who
lookedtired.Helookedatthejunior,wholookedterried.Helookedbackatthewhiteboard.
Builditnow.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge7
5WhyYou
TorchLabscan'taordtotakeHirushaoliveoperationslongenoughtobuildProxyMazefrom
scratch.Theycan'thire,atleastnotyet,asecondseniorinfrastructureengineerwhoknows
thesystemwel lenoughtodesignacontinuousmonitoringservicefromablankpage.SoSachin
madeadierentcal l.Hewenttotheuniversities.
ThisdocumentisthebriefthatProxyMazemustbebuiltfrom.YouaretheengineerSachinis
lookingfor.Hehasnoideawhoyouare.Hehasn'tseenyourCV.Hehasn'tinterviewedyou.
Hewil lneverreadyourcode.
Whathewil ldoispointanHTTPclientattheserviceyoubuild,anduseit.
Ifyourserviceanswersthewayitshould,ProxyMazebecomesHirusha'snewsafetynet,your
namegoesonashortlistthatheiskeepinginthatsmal lrulednotebookinhisbackpocket,and
TorchLabstakesasteptowardneverhavingaWednesdaylikethelastoneagain.
5.1WhatProxyMazeMustDo
1.Acceptruntimecongurationformonitoringcadenceandrequesttimeout,andapplyit
immediately.
2.AcceptproxyURLsandmonitorthemcontinuously,inthebackground,onitsown
schedule,notonlywhensomeb o dyasks.
3.Trackeveryproxyaspending,up,ordown,classiedstrictlyfromrealHTTPprob es.
4.Computethep o olfailurerateas
down
total
## .
5.Fireanalertthemomentthefailureratereachesthethreshold,andresolvethatsame
alertwhenthep o olrecoversb elowthethreshold.
6.Deliveralert.firedandalert.resolvedwebho okeventstoregisteredreceivers,in-
cludingundertransientfailuresofthereceiver.
7.Keepeveryendp ointandeverywebho okpayloadtellingexactlythesamestoryab out
exactlythesamemonitoringstate.
8.Sp eakSlackandDiscord,sothep eoplewhoactuallyrunTorchLabs(theoneswith
phones,notlogreaders)getausefulmessageinsteadofrawJSON.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge8
6ScoreSummary
CategoryPoints
Serviceb o otstrapandconguration10
Proxyp o olingestionandbackgroundmonitoring45
Singlefailureb ehavior30
Thresholdbreachalertsandwebho okdelivery90
## Alertresolution20
## Re-breachlifecycleintegrity30
Po olop erationsandobservability25
CoreTotal250
Slackintegration(b onus)+10
Discordintegration(b onus)+10
MaximumScore270
PassingScore186
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge9
7ProxyIdentiersandGroundRules
ProxyIDsmustb edeterministicandexternallyaddressable.ForaproxyURLwhose
pathendsinanalsegment,thatsegmentistheproxyid:
https://proxy-provider.example/proxy/px-101
## ↓
px-101
ThesameIDmustapp earconsistentlyacrossPOST           /proxies,GET           /proxies,GET           /proxies/{id},
GET           /proxies/{id}/history,thefailed_proxy_idseldonalerts,andeverywebho okpay-
loadthatmentionsafailingproxy.
AlltimestampsareISO8601UTC,e.g.2026-04-24T10:15:30Z.Allrequestandresp onse
b o diesareJSONunlessotherwisestated.Thethresholdforthep o olfailurerateis0.20.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge10
## PARTTWO
TheOp erationalContract
## Sachinslidaprinteddocumentacrossthetabletotheteam,andacrossthisdocumenttoyou.
Thirteenendpoints.EachoneapieceofthesystemthatHirushahadneededat11:47PMand
didnothave.Eachonewil lbeexercised,inwaysthatarenotannouncedinadvance.
Thechaptersthatfol lowdenewhatProxyMazemustdo.Nothow.Thehowisyourjob.
CHAPTER01GET            /health
Pro ofofLife
## Beforeyoucanmonitoranything,provethatyouexist.
Returnsservicehealth.Resp onse200           OK:
## {
## "status":             "ok"
## }
CHAPTER02POST            /config
TheHeartb eat
Setthepace.Everywatchmanhasafrequency.
Setstheruntimemonitoringconguration.Thevaluesmustapplytoallsubsequenthealth
checks,immediately.Resp onse200           OK.
## {
## "check_interval_seconds":             15,
## "request_timeout_ms":             3000
## }
CHAPTER03GET            /config
TheMemory
Repeatbackwhatyouheard.Trust,butverify.
Returnsthecurrentlyactiveruntimeconguration.Resp onse200           OK,b o dymatchesthe
mostrecentlyacceptedPOST           /config.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge11
CHAPTER04POST            /proxies
BuildingthePo ol
Loadthetargets.Thewatchbeginsthemomentyouacceptthem.
LoadsproxyURLsintothemonitoringp o ol.
## Request:
## {
## "proxies":             [
## "https://proxy-provider.example/proxy/px-101",
## "https://proxy-provider.example/proxy/px-102"
## ],
"replace":             true
## }
## Rules:
replaceomittedorfalse:app endtheprovidedproxiestothecurrentp o ol.
replace:                      true:clearthecurrentp o olrst,thenloadtheprovidedproxies.
Newlyacceptedproxiesstartaspendinguntiltheirrstcheckcompletes.Theymust
transitiontoupordownontheirown,fromrealprob es,withoutb eingasked.
Replacingorclearingthep o olmustnotdeletepreviousalerts.
Additionaleldsnotlistedab ovemayapp earintherequestb o dy.Theymustb eignored
cleanly:theymustnotcausetherequesttofail.
Resp onse201           Created:
## {
## "accepted":             2,
## "proxies":             [
{             "id":             "px-101",             "url":             "...",             "status":             "pending"             },
{             "id":             "px-102",             "url":             "...",             "status":             "pending"             }
## ]
## }
CHAPTER05GET            /proxies
TheWatchtower
## Surveytheentirepoolataglance.
Returnsthelivep o olsummaryandp er-proxystate.Resp onse200           OK:
## {
## "total":             10,
## "up":             7,
## "down":             3,
## "failure_rate":             0.3,
## "proxies":             [
## {
## "id":             "px-101",
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge12
## "url":             "https://proxy-provider.example/proxy/px-101",
## "status":             "up",
"last_checked_at":             "2026-04-24T10:15:30Z",
## "consecutive_failures":             0
## }
## ]
## }
Eachproxyentrymustcarry,atminimum:id,url,status,last_checked_at,consecutive_failures.
## Thevaluesmustreectthelatestbackgroundcheck,notacheckfreshlytriggeredbythisre-
quest.
CHAPTER06GET            /proxies/{id}
TheDossier
Everyproxyhasastory.Thisiswhereyoureadit.
Returnsdetailsforasingleproxy.Return404           Not           FoundforunknownIDs.Resp onse200
## OK:
## {
## "id":             "px-101",
## "url":             "https://proxy-provider.example/proxy/px-101",
## "status":             "up",
"last_checked_at":             "2026-04-24T10:15:30Z",
## "consecutive_failures":             0,
## "total_checks":             12,
## "uptime_percentage":             91.7,
## "history":             [
{             "checked_at":             "2026-04-24T10:15:30Z",             "status":             "up"             }
## ]
## }
ThefullsetofrequiredeldsisthevefromGET           /proxies,plustotal_checks,uptime_percentage,
andhistory.
CHAPTER07GET            /proxies/{id}/history
TheChronicle
## Thepapertrailthatproveswhathappened,andwhen.
Returnsthecheckhistoryforasingleproxy.Theresp onseb o dymustb eaJSONarray.
Return404           Not           FoundforunknownIDs.Resp onse200           OK:
## [
{             "checked_at":             "2026-04-24T10:15:30Z",             "status":             "up"                                        },
{             "checked_at":             "2026-04-24T10:16:00Z",             "status":             "down"             }
## ]
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge13
CHAPTER08DELETE            /proxies
TheGraveyard
Cleartheboard.Historysurvivesthepurge.
Clearsthecurrentproxyp o ol.Resp onse:204           No           Content.
Theproxyp o olb ecomesempty.
ExistingalertsmustremainaccessiblethroughGET           /alerts.
Alerthistorymustnotb edeleted.
CHAPTER09GET            /alerts
TheAlertArchive
Everycrisis,recorded.Activeandresolved,nothingforgotten.
Returnsallalerts,b othactiveandresolved.Resp onse200           OK,b o dyisaJSONarray:
## [
## {
## "alert_id":             "alert-a1b2c3",
## "status":             "active",
## "failure_rate":             0.3,
## "total_proxies":             10,
## "failed_proxies":             3,
"failed_proxy_ids":             ["px-103",             "px-104",             "px-105"],
## "threshold":             0.2,
"fired_at":             "2026-04-24T10:20:00Z",
"resolved_at":             null,
"message":             "Proxy             pool             failure             rate             exceeded             threshold"
## }
## ]
7.0.1Alertob ject:requiredelds
alert_id:non-empty,stableforthelifetimeofthealert.Anewalert_idismintedon
everyfreshbreach.
## status:"active"whilethebreachholds,"resolved"afterrecovery.
## failure_rate:theratethatjustiedthealert(≥0.20).
total_proxies:thesizeofthep o olatretime.
## failed_proxies:countofproxiescurrentlydown.
failed_proxy_ids:IDsoftheproxiescurrentlydown.
## threshold:0.2.
fired_at:ISO8601UTC,themomentthebreachb egan.
resolved_at:ISO8601UTConceresolved,otherwisenull.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge14
## message:ashort,non-emptyhuman-readablesummary.
Alertlifecycle.Analertisalifecycle,notacounter.Atmostonealertisactiveatany
time.Whilethebreachcontinues,thatsinglealertanditssinglealert_idmustp ersist:
nosecondactivealertmayapp ear,nomatterhowlongthebreachlasts.Afterresolution,
afreshbreachmintsabrand-newalert_id;thepreviouslyresolvedalertremainsinthe
archiveunchanged.
CHAPTER10POST            /webhooks
TheMessenger
## Whenthealarmres,somebodygetswokenup.
RegistersaURLtoreceivealertwebho oknotications.Therequestb o dyisaJSONob ject.
Theurleldisrequired;additionaleldsmayb epresentandmustb eaccepted(and
ignored)withouterror.
## Requestexample:
## {
## "url":             "https://receiver.example/proxywatch-webhook"
## }
Resp onse201           Created:
## {
## "webhook_id":             "wh-123",
## "url":             "https://receiver.example/proxywatch-webhook"
## }
## 7.0.2alert.firedpayload
## {
## "event":             "alert.fired",
## "alert_id":             "alert-a1b2c3",
"fired_at":             "2026-04-24T10:20:00Z",
## "failure_rate":             0.3,
## "total_proxies":             10,
## "failed_proxies":             3,
"failed_proxy_ids":             ["px-103",             "px-104",             "px-105"],
## "threshold":             0.2,
"message":             "Proxy             pool             failure             rate             exceeded             threshold"
## }
## 7.0.3alert.resolvedpayload
## {
## "event":             "alert.resolved",
## "alert_id":             "alert-a1b2c3",
"resolved_at":             "2026-04-24T10:30:00Z"
## }
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge15
7.0.4Deliveryrequirements
SendeacheventwithheaderContent-Type:                      application/json.
Delivereacheventtoeveryregisteredreceiverwithin60secondsoftheunderlyingstate
transition.
Ifthereceiverresp ondswithatransientfailure(500,502,503,or504),retryuntilthe
deliverysucceeds.
Foreachstatetransition,exactlyonesuccessfuldeliverymustreachthereceiver.No
duplicates,evenifthebreachp ersistsforanextendedp erio d.
CHAPTER11POST            /integrations
TheIntegrationLayer
Speakthelanguageoftheteam.Slackforops.Discordforengineering.
RegistersaSlackorDiscordformattedalertintegration.Resp onse:200           OKor201           Created.
Detailedpayloadrequirementsapp earinPartFour:BonusIntegrations.
## Slackrequest:
## {
## "type":             "slack",
## "webhook_url":             "https://receiver.example/slack",
"username":             "ProxyWatch",
## "events":             ["alert.fired",             "alert.resolved"]
## }
## Discordrequest:
## {
## "type":             "discord",
## "webhook_url":             "https://receiver.example/discord",
"username":             "ProxyWatch",
## "events":             ["alert.fired",             "alert.resolved"]
## }
CHAPTER12GET            /metrics
TheControlRo om
## Whatgetsmeasured,getsmanaged.
Returnsop erationalmonitoringdata.Theresp onseb o dymustb evalid,non-emptyJSON.
Resp onse200           OK:
## {
## "total_checks":             120,
## "current_pool_size":             10,
## "active_alerts":             1,
## "total_alerts":             3,
## "webhook_deliveries":             4
## }
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge16
## PARTTHREE
TheLawsofProxyMaze
Everysystemhasrules.Thesearenotsuggestions.Theyarethebehaviouralconstraintsthat
determinewhetherProxyMazeactual lysolvestheproblemSachinwroteonthewhiteboard.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge17
8BehavioralRules
▶Monitoringrunscontinuously,inthebackground,onthecadencesetby
check_interval_seconds.Aservicethatonlychecksproxieswhenareadendp oint
iscalleddo esnotsatisfytherequirement.
▶ProxystatusisderivedfromrealHTTPprob estothesubmittedproxyURLs,
neverfrommo cked,hardco ded,orcacheddata.
▶A2xxresp onsereceivedwithinrequest_timeout_ms⇒up.
▶Atimeout,connectionfailure,connectionrefusal,orany5xxresp onse⇒down.
## Bothtimeout-stylefailuresand5xx-stylefailuresmustclassifytheproxyasdown;
b othmustapp earinfailed_proxy_idsduringabreach.
▶Thealertthresholdis0.20.Thealertreswhenthefailurerateis≥0.20andresolves
whenthefailureratedropsb elow0.20.
▶Onlyonealertisactiveatatime.Acontinuousbreachmustnotpro duceduplicate
activealertsandmustnotpro duceduplicatesuccessfulalert.fireddeliveriestoa
registeredreceiver.
▶Afteranalertresolves,afreshbreachcreatesanewalertwithabrand-new
alert_id.Thepreviouslyresolvedalertstaysinthearchive.
▶Foreachtransition,thewebho okeventsmustb eobservableinorder:alert.fired
fortheprioralert,thenalert.resolvedfortheprioralert,thenalert.firedfor
thenewalert.
## ▶failed_proxy_idsmustalwaysequalthesetofproxiescurrentlyclassiedasdown.
▶GET           /proxies,GET           /alerts,andanywebho okpayloaddescribinganactivebreach
mustagreeexactlyonthefailedproxyset,ontotal_proxies,onfailed_proxies,
andonthreshold.
▶Allrequestb o diesthatcontainaJSONob jectmustacceptunknowneldswithout
error.Rejectonlyongenuinelymalformedinput.
▶SampleproxyURLs,IDs,orderings,timings,andfailuremo desusedduringevalua-
tionmaydierfromanyexamplesinthisdo cument.Thecontractiswhatmatters,
notthesp ecicexamples.
8.1AlertLifecycle
NormalActiveAlertResolvedNewAlert
rate≥0.20
rate<0.20
rate≥0.20
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge18
## PARTFOUR
BonusIntegrations
ProxyMaze'scoreworkswithoutthese.Butthepeoplewhoactual lyrunTorchLabs(Hirusha,
thecustomersuccessteam,theengineersinthe#nocchannel)liveinSlackandDiscord.A
systemthatspeakstheirlanguagegetsused.Onethatdoesn'tgetsignored,nomatterhowwel l
itworksunderneath.
BONUSINTEGRATION01Slack|+10pts
## SLACKFORMATTEDALERTS
Oneveryalert.fired(andalert.resolved)event,POSTaSlack-stylepayloadtothe
registeredwebhook_url,withContent-Type:                      application/json.Thepayloadmust
contain:
## username:non-emptystring.
## text:non-emptystringsummarisingtheevent.
attachments[0].color:ahexstringoftheform"#RRGGBB".
attachments[0].fields:arrayof{title,           value}entries.Thetitlescollectively
mustinclude(case-insensitivesubstringmatch):AlertID,FailureRate,FailedProx-
ies,Threshold,FailedIDs,andFiredAt.
## attachments[0].footer:non-emptystring.
attachments[0].ts:aUnixep o chtimestampexpressedasaninteger(numb erof
wholeseconds;notaoatandnotastring).
TheSlackpayloadmustb edeliveredtotheregisteredSlackwebho okURLwithin60
secondsoftheunderlyingalertevent.
ProxyMaze'26CondentialPoweredby

ProxyMaze'26TheTorchLabsChallenge19
BONUSINTEGRATION02Discord|+10pts
## DISCORDFORMATTEDALERTS
Oneveryalertevent,POSTaDiscord-stylepayloadtotheregisteredwebhook_url,with
Content-Type:                      application/json.Thepayloadmustcontain:
## embeds[0].title:non-emptystring.
## embeds[0].description:non-emptystringsummarisingtheevent.
## embeds[0].color:anintegerintherange0to16777215inclusive.
embeds[0].fields:arrayof{name,           value}entries.Thenamescollectivelymust
include(case-insensitivesubstringmatch):AlertID,FailureRate,FailedProxies,
Threshold,andFailedIDs.
## embeds[0].footer.text:non-emptystring.
TheDiscordpayloadmustb edeliveredtotheregisteredDiscordwebho okURLwithin
## 60secondsoftheunderlyingalertevent.
Go o dluck.
Hirushaiswatchingthedashb oard.
Sachiniswatchingthescoreb oard.
SriLankaiswatchingyou.
TorchLabs•Colomb o,SriLanka•FromSriLanka,totheworld.
ProxyMaze'26CondentialPoweredby