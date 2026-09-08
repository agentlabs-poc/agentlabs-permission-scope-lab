/* Editorial artifact generator. Not an Auth implementation or schema validator. */
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const pptxgen = require('pptxgenjs');

const OUT = __dirname;
const W = 16, H = 9;
const C = { bg:'F5F7FA', ink:'14263D', muted:'52657A', teal:'087E8B', pale:'E3F3F2', code:'14263D', codeText:'ECF3FA', codeMuted:'A3B2C3', purple:'6945B8', line:'CCD7E2', amber:'9A5A06', amberBg:'FFF2D9', white:'FFFFFF', red:'AC334A' };
const pptx = new pptxgen();
pptx.defineLayout({name:'AUTH_WIDE',width:W,height:H});
pptx.layout = 'AUTH_WIDE';
pptx.author = 'Agentlabs';
pptx.subject = 'Handbook of Authorization — core vocabulary and approved JSON';
pptx.title = 'Handbook of Authorization — Core Concepts & JSON';
pptx.company = 'Agentlabs';
pptx.lang = 'en-IN';
pptx.theme = {headFontFace:'Arial',bodyFontFace:'Arial',lang:'en-IN'};

const REF = '../theory/canonical-terms.md';
const READ = 'hrms:employee:certificate::read';
const WRITE = 'hrms:employee:certificate::write';
const G1 = {version:'1',grant_id:'G1',revision:2,parent_grant_id:'G0',permissions:[READ,WRITE],scope:{dept:'FIN'}};
const G2 = {version:'1',grant_id:'G2',revision:1,parent_grant_id:'G1',permissions:[READ],scope:{cert:'C17'}};
const A1 = {version:'1',id:'A1',grant_id:'G1',grant_revision:2,recipient:{type:'group',id:'Team1'},status:'enabled'};
const A2 = {version:'1',id:'A2',grant_id:'G2',grant_revision:1,recipient:{type:'group',id:'Team2'},status:'enabled'};
const roleGrant = {version:'1',grant_id:'G-ROLE-READER',revision:1,parent_grant_id:'G0',role_id:'R-CERTIFICATE-READER',role_revision:1,scope:{dept:'FIN'}};
const identity = (type,id) => ({version:'1',actor:{type,id},human_id:'U-17'});
const getPolicy = {version:'1',method:'GET',path:'/api/v1/{tenant}/{dept}/{cert}',permission:READ,inputs:{tenant:{source:'path',name:'tenant'},dept:{source:'path',name:'dept'},cert:{source:'path',name:'cert'}}};
const putPolicy = {version:'1',method:'PUT',path:'/api/v1/{tenant}/certificates/{cert}',permission:WRITE,inputs:{tenant:{source:'path',name:'tenant'},cert:{source:'path',name:'cert'},proposed_dept:{source:'body',name:'department_id'}}};
const allow = {version:'1',decision:'allow',grant_ids:['G2']};
const deny = {version:'1',decision:'deny',error_code:'NO_AUTHORIZING_GRANT',error_message:'You do not have access.',error_message_reason:'No complete route permits this read.'};
const error = {version:'1',error_code:'AUTH_SERVICE_TIMEOUT',error_message:'We could not check your access.',error_message_reason:'Required authority could not be loaded.'};

// Small nested values stay compact; complete JSON remains parseable.
function pretty(obj) {
  return JSON.stringify(obj,null,2).replace(/\{\n\s+"([^"\n]+)": "([^"\n]+)"(?:,\n\s+"([^"\n]+)": "([^"\n]+)")?\n\s+\}/g,
    (_,k,v,k2,v2)=>`{"${k}": "${v}"${k2?`, "${k2}": "${v2}"`:''}}`);
}
const modelNotes = 'These are the repository model’s choices, not universal authorization requirements. Canonical core examples are not complete deployable schemas. Tenant context, registered meanings, live controls, valid assignments and required upstream support are premises of usable authority. Revision/adoption mechanics are intentionally outside this deck; existing revision fields remain for contract fidelity.';
const slides = [];
const add = (d) => slides.push(d);
const pair = (term,text) => ({term,text});

add({kind:'cover',title:'Handbook of\nAuthorization',subtitle:'Core concepts & JSON',section:'FOUNDATIONS',takeaway:'Vocabulary first. Relationships second. JSON beside the concept.',notes:'PPT-001: approved vocabulary-first presentation. The user asked to keep revision mechanics aside. This deck teaches definitions and approved representations without changing any canonical contracts. Presenter notes retain qualifications that do not need to dominate the slides.',source:REF});

add({kind:'cards',section:'THE VOCABULARY',title:'Six questions make the model readable',subtitle:'Each term answers a different question.',cards:[pair('Identity','Who is acting, and whose human authority supports the request?'),pair('Permission','What operation can be performed?'),pair('Scope','Within which boundary?'),pair('Grant','Which bounded authority is defined?'),pair('Assignment','Who receives that grant?'),pair('Evaluation','Does valid authority cover this request?')],takeaway:'A grant defines authority. An assignment connects it to a recipient.',notes:'Authentication establishes identity. Authorization connects authority to a protected effect. A principle is a design constraint; a policy expresses applicable rules. Effective authority is what current valid routes can provide. In all JSON, version identifies the format. Revision fields are retained but visually muted and not taught in this deck. No revision/adoption lesson is included.',source:REF+'#principles-and-rules'});

add({kind:'split',section:'IDENTITY',title:'Human, actor and principal',subtitle:'The actual caller and the authority anchor are distinct roles.',points:[pair('Human / user','The human identity; actor type is user.'),pair('Actor / caller','The person or program making this request.'),pair('Principal','An umbrella term—not another record or field.')],code:identity('user','U-17'),codeLabel:'APPROVED CORE JSON · DIRECT HUMAN',takeaway:'For a direct human: actor.id equals human_id.',notes:'Identity describes who acts. human_id identifies the authorizing human, anchoring current authority and self. Direct-human equality is required. Matching JSON is not identity proof: the identity must be established through trusted evidence. The typed actor avoids conflating a proxy caller with the human.',source:REF+'#identity'});

add({kind:'split',section:'IDENTITY',title:'Agent and service account are proxies',subtitle:'In our model, neither supplies independent authority.',points:[pair('Actor','The actual automated caller.'),pair('human_id','The human whose authority bounds the request.'),pair('Trust required','The association and delegation must be established.')],codes:[{label:'AGENT · APPROVED IDENTITY SHAPE',value:identity('agent','A-17')},{label:'SERVICE ACCOUNT · SAME SHAPE',value:identity('service_account','SA-17')}],takeaway:'Automated actors are not first-class members of authorization teams.',notes:'The human-dependent service-account rule is specific to this framework. Other systems may choose independent machine principals. Neither JSON block is a delegation record: both identify the participants. Exact delegation evidence remains pending. There is no separate duplicate agent_id claim.',source:REF+'#identity'});

add({kind:'split',section:'IDENTITY',title:'JWT subject stays human',subtitle:'Subject identifies the human; actor identifies the caller.',points:[pair('sub','The human subject in our JWT profile.'),pair('identity.actor','The actual caller, here A-17.'),pair('Same identity block','Reusable beyond the JWT transport.')],code:{version:'1',sub:'U-17',identity:identity('agent','A-17')},codeLabel:'JWT PAYLOAD EXCERPT · NOT A COMPLETE TOKEN',takeaway:'Keeping sub compatible does not bypass proxy restrictions.',notes:'This identity-related payload excerpt omits issuer, audience, expiry and cryptographic verification for focus, not because they are optional. sub must equal identity.human_id. The repeated human ID keeps identity self-contained. Legacy consumers that ignore the proxy must not bypass applicable delegation checks. Universal backward compatibility has not been demonstrated.',source:REF+'#identity'});

add({kind:'cards',section:'CONTEXT',title:'Tenant is the enclosing boundary',subtitle:'A path claim is not trusted tenant context.',cards:[pair('Tenant context','Established outer isolation for ordinary tenant authority.'),pair('Scope','A narrower selection inside the enclosing boundary.'),pair('Application','Owns the meanings of its operations and boundary keys.')],takeaway:'Tenant is implied in grant scope—not a removable scope key.',notes:'A tenant identifier in a URL is supplied input, not proof of access. Tenant association is not authorization-group membership. Application platform administration governs capability publication outside tenant business scope, within its own management boundary. Publishing HRMS permissions does not grant tenant payroll access. Complete context and platform trust mappings remain pending.',source:REF+'#enclosing-context'});

add({kind:'split',section:'AUTHORITY',title:'Permission names the operation',subtitle:'What can be done—not who can do it or how far it reaches.',points:[pair('Namespace','Application, domain and optional deeper nouns.'),pair('Verb','The operation after the double colon.'),pair('No inference','No prefix inheritance, alias or wildcard access.')],raw:'[\n  "hrms:employee:certificate::read",\n  "hrms:employee:certificate::write"\n]',fragment:true,codeLabel:'NESTED permissions ARRAY · NOT A GRANT',takeaway:'One endpoint requires one permission. One grant may contain many.',notes:'Permission grammar: <app>:<domain>[:<subdomain-or-resource>...]::<verb>. Permissions must be registered. Department IDs do not belong in operation names. The array is a nested value, not a standalone published contract, so it does not invent a version wrapper. An existing permission name cannot acquire materially different authorization meaning after retirement. HTTP method alone does not define permission.',source:REF+'#permission'});

add({kind:'split',section:'AUTHORITY',title:'Scope selects the boundary',subtitle:'The operation must stay inside every applicable constraint.',points:[pair('Keys','Application-registered boundary dimensions.'),pair('Values','Non-empty strings with application-defined meaning.'),pair('Combination','Entries within a scope combine with AND.')],raw:'{\n  "dept": "FIN",\n  "cert": "C17"\n}',fragment:true,codeLabel:'NESTED scope VALUE · NOT A STANDALONE RECORD',takeaway:'This means Finance AND C17—not Finance OR C17.',notes:'Our scope is a required flat object. Arrays, nested query objects, duplicate keys, wildcard operators, unsupported keys and empty/non-string values are not accepted. Missing/null scope is not an empty object. A child’s {} adds no local restriction but retains parent constraints. Alternative boundaries need separate complete routes. Auth validates registration and syntax; the application owns boundary meaning.',source:REF+'#scope-and-self'});

add({kind:'split',section:'AUTHORITY',title:'$self means the authorizing human',subtitle:'A group can receive self-scoped authority without becoming “self”.',points:[pair('Group recipient','Employees receives the grant through an assignment.'),pair('Vinay’s request','Self means Vinay, not Employees or its owner.'),pair('Vinay’s agent','Self remains Vinay, with delegation limits.')],raw:'{\n  "user": "$self"\n}',fragment:true,codeLabel:'NESTED scope VALUE · REGISTERED SELF RELATIONSHIP',takeaway:'Runtime self is settled; cross-recipient issuance needs separate validation.',notes:'This example assumes an already valid supported self-scoped group route. It does not settle whether Maya’s self-limited source permits distributing Nutan’s self access. Copying the token is not a containment proof. The application must establish or enforce its registered user/self relationship on the actual data. Prefer group-based access without forbidding legitimate direct assignments.',source:REF+'#scope-and-self'});

add({kind:'cards',section:'ROLE · GRANT · ASSIGNMENT',title:'Three concepts—not three names for access',subtitle:'Keep the permission bundle, bounded authority and recipient separate.',cards:[pair('ROLE','A reusable bundle of permissions.\nExample: certificate read + write.'),pair('GRANT','Permissions, scope and required support.\nExample: editor authority within Finance.'),pair('ASSIGNMENT','Connects a grant to a recipient.\nExample: Team1 receives G1.')],takeaway:'Roles are optional. A grant may list permissions directly.',notes:'A role does not define recipient or scope. A grant uses explicit permissions or a complete role reference, not both. An assignment connects a reusable recipient-free grant to a human or group. The full standalone role-publication schema remains pending; this conceptual comparison deliberately does not invent one. Scope and parent support still constrain role-based authority.',source:REF+'#grant-records'});

add({kind:'split',section:'GRANT',title:'A grant defines bounded authority',subtitle:'Finance certificate read/write, dependent on valid G0 support.',points:[pair('permissions','The operations selected from supported authority.'),pair('scope','The local Finance boundary.'),pair('parent_grant_id','The required upstream grant—not proof of support.')],code:G1,codeLabel:'APPROVED GRANT-CONTENT SHAPE · REVISION FIELDS MUTED',takeaway:'No recipient belongs on the reusable grant.',notes:'G1 matches the handbook running example. Assume registered meanings, valid G0 support, enabled live control and successful administrative/boundary checks. The example is immutable content, not a complete combined record: live control and recipient assignment are separate. Revision mechanics are omitted from teaching, but required fields remain intact. Ordinary grants cannot omit their parent to manufacture roots.',source:REF+'#grant-records'});

add({kind:'split',section:'ROLE',title:'A grant can use a role instead',subtitle:'The role supplies the permission bundle; the grant supplies its boundary.',points:[pair('role_id','The reusable bundle, here Certificate Reader.'),pair('scope','Finance remains a grant boundary.'),pair('Exclusive forms','Use permissions OR the complete role reference.')],code:roleGrant,codeLabel:'APPROVED ROLE-REFERENCE GRANT SHAPE',takeaway:'A role is not a team, an assignment or an unrestricted grant.',notes:'Assume R-CERTIFICATE-READER contains certificate-read and fits valid G0 support. Full standalone role publication JSON is not approved by showing this reference. role_revision is preserved in the canonical JSON but its mechanics are outside scope. A direct-permission grant contains neither role_id nor role_revision. No latest-role fallback or mixed variant is allowed.',source:REF+'#grant-records'});

add({kind:'split',section:'ASSIGNMENT',title:'An assignment names the recipient',subtitle:'Grant G1 is supplied to Team1 through A1.',points:[pair('id','The assignment identity, not the grant ID.'),pair('grant_id','Which reusable grant is assigned.'),pair('recipient','Who receives it: a human or group.'),pair('status','This assignment’s own live control.')],code:A1,codeLabel:'APPROVED ASSIGNMENT SHAPE',takeaway:'Creating a grant is not assigning it. Assigning it is not membership.',notes:'The canonical assignment also retains its required content-selection field without teaching adoption mechanics. Prefer human access through groups while retaining valid direct-human routes. Assignment enablement is not grant-wide enablement or proof of effective authority. An agent does not get independent first-class group membership through this record.',source:REF+'#grant-records'});

add({kind:'cards',section:'TEAMS & PEOPLE',title:'Membership is not ownership',subtitle:'Team and group mean the same thing in our model.',cards:[pair('MEMBERSHIP','Nutan is a Team2 member.\nShe can use its effective assigned authority.'),pair('ASSIGNMENT','A2 supplies G2 to Team2.\nIt does not add Nutan as a member.'),pair('OWNERSHIP / ADMIN','Maya can perform explicitly authorized management.\nThat does not imply business access.')],takeaway:'Team2 membership does not automatically create Team1 membership.',notes:'Auth owns explicit human membership. Team hierarchy does not imply inherited human memberships. Teams are not application departments even when used to organize department staff. Team write includes membership management; assigning grants needs separate authority. Full membership, team hierarchy and owner record formats remain pending, so no tentative scratch JSON is promoted. Authorized membership management intentionally distributes the team’s existing access.',source:REF+'#teams-and-administration'});

add({kind:'gates',section:'ADMINISTRATION',title:'Can assign? And may assign this authority?',subtitle:'Both checks happen inside Auth Service for an authority-changing request.',takeaway:'Administrative permission alone is not proof that a proposed grant fits.',notes:'The Auth endpoint evaluator checks the administrative operation and recipient boundary. Auth’s authority-boundary validator checks the proposed authority against eligible source support and team ceilings. Q-093 requires a valid direct/group source available to the assigner as well as administrative authority. This is not an application database lookup or a generic business-rule engine. Creator history is not automatically permanent support. Exact owner-transfer permission remains pending.',source:REF+'#responsibility-layers'});

add({kind:'split',section:'DEPENDENT AUTHORITY',title:'A child grant can only narrow',subtitle:'G2 selects read and adds C17 to G1’s Finance boundary.',points:[pair('Permission subset','Read is selected from parent read/write.'),pair('Scope AND','Effective scope is Finance AND C17.'),pair('Live dependency','The parent definition alone is not support.')],code:G2,codeLabel:'APPROVED CHILD GRANT-CONTENT SHAPE',takeaway:'Child {} retains the parent boundary. It never becomes an escape hatch.',notes:'Assume Team2 is a child of Team1, A1 supplies G1 to Team1, and A2 supplies G2 to Team2 with valid controls/support. All child-team authority must remain within the parent-team ceiling as well. Adding dept=ENG cannot overwrite dept=FIN. Subgrant means a grant relative to a parent, not another record type. Full team hierarchy wire representation remains pending.',source:REF+'#dependent-relationships'});

add({kind:'lineage',section:'DEPENDENCY GRAPH',title:'Two lineages, joined by assignments',subtitle:'Team relationships and grant relationships must both be established.',takeaway:'Team lineage is not grant lineage; neither implies human membership.',notes:'The graph repeats G1/A1/Team1 and G2/A2/Team2 from the handbook. Required G0 support is assumed. G2 parent_grant_id is G1; A1 establishes actual support in the parent-team context. No parent_assignment_id or parent_grant_revision field is introduced. Nutan’s explicit Team2 membership supplies the valid G2 route, not Team1 write access. Unrelated TeamX holdings cannot replace required Team1 support. Complete team relationship JSON is pending.',source:REF+'#dependent-relationships'});

add({kind:'cards',section:'DELEGATION',title:'Proxy authority stays within the human',subtitle:'Identity names the participants; delegation establishes the dependent authority.',cards:[pair('HUMAN CEILING','Vinay currently has read + write.'),pair('DELEGATION LIMIT','A-17 is permitted read only.'),pair('EFFECTIVE PROXY','A-17 can read, not write.\nLosing Vinay’s read support removes that route.')],takeaway:'Direct human-to-proxy delegation only; no proxy-to-proxy chains in v1.',notes:'Delegation is not simply an account creator field. Both current human authority and delegation limits apply. If support later returns, still-valid delegation can resume; expired, revoked or explicitly disabled delegation is not automatically revived. Full delegation evidence JSON is pending. Teams contain humans, not these automated actors.',source:REF+'#delegation'});

add({kind:'split',section:'LIFECYCLE',title:'Enabled does not mean effective',subtitle:'Assignment and grant controls answer different questions.',points:[pair('Grant control','Disabling G2 withdraws all routes through G2.'),pair('Assignment control','Disabling A2 affects that recipient binding.'),pair('Effective authority','Controls, support and constraints must all pass.')],codes:[{label:'GRANT-WIDE CONTROL',value:{version:'1',id:'G2',status:'disabled'}},{label:'ASSIGNMENT CONTROL',value:{...A2,status:'enabled'}}],takeaway:'An enabled assignment cannot override a disabled grant.',notes:'This slide is a later lifecycle snapshot, not a contradiction of the earlier valid-route example. Stored enabled flags are necessary controls, not final authorization. A still-enabled child can become ineffective when required support is disabled and resume when that support returns. A child explicitly disabled stays disabled until explicitly enabled and revalidated. Structural changes require bottom-up handling of affected bindings, not merely making an ancestor ineffective.',source:REF+'#lifecycle'});

add({kind:'split',section:'LIFECYCLE',title:'Validity limits when authority applies',subtitle:'An expiry is a time boundary—not a new permission or scope key.',points:[pair('expires_at','The exclusive end of the validity window.'),pair('not_before','An optional inclusive start.'),pair('Parent limits','A local window cannot outlive required support.')],code:{version:'1',grant_id:'G-TEMP-READER',revision:1,parent_grant_id:'G1',permissions:[READ],scope:{cert:'C17'},validity:{expires_at:'2026-09-30T00:00:00Z'}},codeLabel:'APPROVED GRANT VALIDITY PLACEMENT',takeaway:'Re-enabling does not reset expiry. Assignment-specific windows are deferred.',notes:'Optional validity belongs in immutable grant content, not live grant control or assignment. The full timestamp and lifecycle API contracts remain unfinished. Absence of a local validity window means no extra local time restriction, not independence from upstream expiry. This deck does not teach content revision or adoption mechanics, but preserves approved field placement.',source:REF+'#lifecycle'});

add({kind:'cards',section:'LIFECYCLE',title:'An orphan has lost required support',subtitle:'Orphaning describes an affected route—not an invented status field.',cards:[pair('ORPHAN','Required parent support no longer exists.\nThat lineage and dependent descendants are ineffective.'),pair('NOT AN ORPHAN','A legitimate root; an unassigned definition; merely disabled or expired support.'),pair('NOT PROVEN','An Auth timeout means evidence is unavailable.\nIt does not prove that support was deleted.')],takeaway:'No automatic deletion, new parent, or authority repair.',notes:'Orphaning is assessed for the affected lineage of a reusable grant, not automatically every assignment of that definition. Permanent deletion/revocation differs from temporary disablement; no invented revoked/orphan live status is selected. Required structural guards still apply before breaking bindings. Missing support cannot be replaced with any convenient broader source. Warning/prevention at an upper management layer does not alter the canonical ineffective outcome.',source:REF+'#lifecycle'});

add({kind:'flow',section:'REGISTRATION & ROOTS',title:'Register meanings, then establish authority',subtitle:'A catalog is not a grant, and registration is not business access.',steps:[pair('REGISTER','Application permissions + scope definitions'),pair('VALIDATE','Syntax, declared compatibility and authority bounds'),pair('BOOTSTRAP','Trusted root + administrator group + human membership'),pair('DISTRIBUTE','Explicit bounded grants and assignments')],takeaway:'Ordinary grants cannot become roots by simply omitting their parent.',notes:'Applications own domain meaning; Auth validates registered contracts and canonical authority rules. Optional permission/scope compatibility is declared upfront; enabled mode validates every grant. Bootstrap must establish coherent maximum intended authority inside its authorized boundary, with a minimal setup. Computed root permissions follow one shared application catalog; ordinary child permissions do not silently grow. Complete registration/root/trusted-setup JSON remains pending; no stored wildcard or is_root field is invented.',source:REF+'#registration'});

add({kind:'split',section:'ENDPOINT POLICY',title:'Declare the required operation and inputs',subtitle:'A server-owned policy connects method + path to one permission.',points:[pair('permission','One registered operation for this endpoint.'),pair('inputs','Selected local names—not the whole request.'),pair('source + name','Exactly where each required value comes from.')],code:getPolicy,codeLabel:'APPROVED GET POLICY CORE SHAPE',takeaway:'A path claim about Finance does not prove C17 belongs to Finance.',notes:'All declared inputs must be present at their stated sources. The tenant path input must be reconciled with trusted tenant context. Local input names are not automatically scope keys. There is no prepared handoff or canonical relationships block. The endpoint must establish or enforce the relationships needed to constrain actual execution. Full policy validation, nested-body syntax and additional sources remain pending.',source:REF+'#endpoint-declaration'});

add({kind:'split',section:'ENDPOINT POLICY',title:'Body field and local input can differ',subtitle:'PUT binds proposed_dept to the body’s department_id.',points:[pair('Local name','proposed_dept is used by this endpoint.'),pair('Declared source','Read department_id from the body.'),pair('Proposal ≠ fact','The new department does not prove the current one.')],code:putPolicy,codeLabel:'APPROVED PUT POLICY CORE SHAPE',takeaway:'No silent default or query fallback for a missing declared body input.',notes:'The body field must be present at the declared source. Application value/type validation is not a new grant-scope field. Current and proposed operation boundaries must be respected; the exact move-grant composition is still pending. A broader {} grant does not waive declared input presence. The one-permission requirement does not mean one grant contains only one permission.',source:REF+'#endpoint-declaration'});

add({kind:'cards',section:'REQUEST VOCABULARY',title:'Request → material → resolved request',subtitle:'Becoming evaluation-ready does not mean becoming authorized.',cards:[pair('REQUEST','Nutan asks to read C17 in Finance.\nGET /api/v1/acme/FIN/C17'),pair('MATERIAL','Declared inputs + verified identity/tenant + trusted facts or enforceable constraints.'),pair('RESOLVED REQUEST','The operation and required meaning/material are established for evaluation.')],takeaway:'Complete request/resolved-request JSON is pending; these are conceptual views.',notes:'A request input is a declared value from a specified source. A domain fact is application-owned knowledge; a relationship is the relevant association, such as C17 belonging to Finance. A supplied path value alone is not a fact. The endpoint may fetch authoritative facts or preserve the required constraints through execution. No canonical target entity, relationship record or resolved_inputs envelope is introduced.',source:REF+'#request-and-material'});

add({kind:'resolved',section:'RESOLUTION',title:'A resolved grant is a dependent view',subtitle:'It preserves the complete route—not an independent copy of permissions.',takeaway:'Resolution cannot amplify authority or mix fragments from unrelated grants.',notes:'Resolved grants are computed views of existing applicable routes, preserving source bindings, scope and restrictions. They are not stored new grants, new assignments or allows. This slide is an explanatory table, not an approved serialization. A FIN-write route and ENG-read route do not provide ENG-write. Independent alternatives remain separate: resolution also does not compute a globally narrowest intersection of every unrelated grant. Full resolved-grant/evidence format remains pending.',source:REF+'#resolution-and-evaluation'});

add({kind:'split',section:'DECISIONS',title:'Allow identifies supporting authority',subtitle:'The result is about this evaluated operation—not future unrestricted access.',points:[pair('decision','A completed allow.'),pair('grant_ids','Non-empty supporting references.'),pair('Evidence','Not all held grants and not a new assignment.')],code:allow,codeLabel:'APPROVED MINIMUM ALLOW RESULT',takeaway:'Supporting references do not replace actual boundary enforcement.',notes:'grant_ids is a non-empty array of non-empty strings. It identifies supporting grants from this evaluation, not necessarily a complete lineage snapshot. Returned scope is not required. References are available even when not every request is recorded for audit; audit storage/retention design is outside scope. Unknown fields and mixed variants are rejected. Full code/transport/provenance schema work remains pending.',source:REF+'#decision-and-enforcement'});

add({kind:'results',section:'DECISIONS',title:'Deny is not an evaluation error',subtitle:'Both stop the protected operation—but they explain different outcomes.',takeaway:'Both message fields reach the UI; the reason is not a private diagnostic channel.',notes:'Deny means sufficient evidence completed evaluation without authorizing the request. Error means a required evaluation could not complete, such as an Auth timeout with no sufficient valid preloaded authority. Error has no decision field, but omission alone does not validate a malformed response. The code strings and message text here are illustrative, not an exhaustive catalog. Both error_message and error_message_reason are evaluator-provided; disclosure and value rules remain pending. Malformed/unknown/mixed result variants cannot become allow.',source:REF+'#decision-and-enforcement'});

add({kind:'flow',section:'ONE CONNECTED REQUEST',title:'From Nutan’s request to a constrained read',subtitle:'Auth supplies authority. The application keeps the operation inside it.',steps:[pair('ESTABLISH','Nutan’s identity + tenant + declared inputs'),pair('RESOLVE','Team2 membership → A2 → G2, supported by Team1 / G1'),pair('EVALUATE','Certificate read inside Finance AND C17'),pair('ENFORCE','Return only data matching the trusted tenant + FIN + C17')],takeaway:'Allow for Finance/C17 must never become an unchecked certificate-ID lookup.',notes:'This is one endpoint-owned gate, not middleware allow followed by an independent prepared decision. Middleware may authenticate and preload sufficient authority. Auth’s canonical layer supplies grants, membership and support; the application layer supplies facts or constrained execution. Inputs appearing only in logs do not enforce boundaries. Already-allowed ordinary synchronous operations and new checks have distinct approved temporal rules; this slide does not invent caching or concurrency protocols.',source:REF+'#responsibility-layers'});

add({kind:'summary',section:'KEEP THE DISTINCTIONS',title:'The vocabulary in one sentence',subtitle:'An identified actor requests an operation; valid assigned authority must cover its boundary.',takeaway:'Definitions first. Approved JSON beside them. Pending formats remain pending.',notes:'End with the canonical distinctions rather than claiming every interface is complete. Sources: handbook/theory/canonical-terms.md; implementation/05-canonical-model.md; implementation/06-auth-service.md; implementation/07-application-integration.md; appendices/pending.md. Revision/adoption mechanics were intentionally left out at the user’s request; required fields remain in examples. Full team/membership/owner, standalone role, registration/root, delegation and request/resolved-view representations remain pending. This deck neither changes the model nor closes those design decisions.',source:REF});

assert.equal(slides.length,30);

// Shared drawing primitives produce editable PowerPoint objects and SVG previews.
const svgs = [], boxes = [], examples = [];
const esc=s=>String(s).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&apos;'}[c]));
let sl, svg, slideNo;
function rect(x,y,w,h,fill,stroke=fill,r=0.12) {
  sl.addShape(r?'roundRect':'rect',{x,y,w,h,rectRadius:r,fill:{color:fill},line:{color:stroke,width:0.7},radius:r});
  svg.push(`<rect x="${x*120}" y="${y*120}" width="${w*120}" height="${h*120}" rx="${r*120}" fill="#${fill}" stroke="#${stroke}"/>`);
}
function wrap(text,width,size,mono=false) {
  const max=Math.floor(width*72/(size*(mono?0.60:0.51)));
  const result=[];
  for(const para of String(text).split('\n')) {
    if(mono){assert(para.length<=max,`Code too wide slide ${slideNo}: ${para.length}>${max}: ${para}`);result.push(para);continue;}
    let line='';for(const word of para.split(' ')){if(line&&line.length+word.length+1>max){result.push(line);line=word;}else line+=(line?' ':'')+word;}result.push(line);
  }
  return result;
}
function text(str,x,y,w,{size=22,color=C.ink,bold=false,mono=false,align='left',line=1.23}={}) {
  const lines=wrap(str,w,size,mono), step=size*line/72, h=lines.length*step+0.08;
  assert(x>=0&&y>=0&&x+w<=W+0.01&&y+h<=H+0.01,`Text overflow slide ${slideNo}: ${str}`);
  const font=mono?'Liberation Mono':'Liberation Sans';
  sl.addText(lines.join('\n'),{x,y,w,h,fontFace:mono?'Courier New':'Arial',fontSize:size,color,bold,margin:0,breakLine:false,paraSpaceAfter:0,lineSpacingMultiple:line,vertAnchor:'top',valign:'top',align,wrap:false});
  lines.forEach((s,i)=>svg.push(`<text x="${(align==='center'?x+w/2:x)*120}" y="${(y+i*step)*120+size*120/72*0.91}" font-family="${font}" font-size="${size*120/72}" font-weight="${bold?700:400}" fill="#${color}" text-anchor="${align==='center'?'middle':'start'}" xml:space="preserve">${esc(s)}</text>`));
  boxes.push({slide:slideNo,text:str,x,y,w,h,size,mono});return h;
}
function arrow(x1,y1,x2,y2,color=C.teal) {
  sl.addShape('line',{x:x1,y:y1,w:x2-x1,h:y2-y1,line:{color,width:2,beginArrowType:'none',endArrowType:'triangle'}});
  svg.push(`<path d="M${x1*120},${y1*120} L${x2*120},${y2*120}" stroke="#${color}" stroke-width="3" fill="none" marker-end="url(#arrow)"/>`);
}
function box(title,body,x,y,w,h,{fill=C.white,accent=C.teal,size=23}={}) {
  rect(x,y,w,h,fill,C.line);rect(x,y,0.055,h,accent,accent,0);
  const th=text(title,x+0.25,y+0.25,w-0.5,{size:18,bold:true,color:accent});
  const bh=text(body,x+0.25,y+0.37+th,w-0.5,{size});
  assert(th+bh+0.55<=h,`Box text overflow slide ${slideNo}: ${title}`);
}
function code(value,label,x=6.35,y=2.55,w=9.0,h=5.25,raw=false) {
  const maxChars=Math.floor((w-0.56)*72/(17*0.60));
  const initial=raw?value:pretty(value);
  const src=initial.split('\n').flatMap(l=>{
    if(l.length<=maxChars)return [l];
    const m=l.match(/^(\s*)("[^"]+":) (.+)$/);
    return m?[m[1]+m[2],m[1]+'  '+m[3]]:[l];
  }).join('\n');
  JSON.parse(src);
  examples.push({slide:slideNo,label,fragment:raw,value:JSON.parse(src)});
  rect(x,y,w,h,C.code);
  text(label,x+0.28,y+0.20,w-0.56,{size:11,bold:true,color:'83D4D6'});
  const lines=src.split('\n');const size=17,step=0.29;
  assert(lines.length*step+0.76<=h,`Code too tall slide ${slideNo}: ${lines.length}`);
  lines.forEach((l,i)=>text(l,x+0.28,y+0.67+i*step,w-0.56,{size,mono:true,line:1.0,color:/"(?:revision|grant_revision|role_revision)"/.test(l)?C.codeMuted:C.codeText}));
}
function points(items) {
  let y=2.63;
  const compact=items.length>3, size=compact?20:22, line=compact?1.1:1.23;
  for(const p of items){const h=text(p.term,0.72,y,5.12,{size,line,bold:true,color:C.teal});y+=h+0.05;y+=text(p.text,0.72,y,5.12,{size,line})+(compact?0.15:0.32);}
  assert(y<8.15,`Points too tall slide ${slideNo}`);
}
function base(d) {
  sl=pptx.addSlide();sl.background={color:C.bg};
  svg=[`<svg xmlns="http://www.w3.org/2000/svg" width="1920" height="1080" viewBox="0 0 1920 1080"><defs><marker id="arrow" markerWidth="10" markerHeight="10" refX="8" refY="5" orient="auto"><path d="M0,0 L10,5 L0,10 Z" fill="#${C.teal}"/></marker></defs><rect width="1920" height="1080" fill="#${C.bg}"/>`];
  rect(0,0,0.14,H,C.teal,C.teal,0);
  text(d.section,0.70,0.42,13.5,{size:13,bold:true,color:C.teal});
  text(d.title,0.67,0.92,14.6,{size:34,bold:true});
  text(d.subtitle,0.70,1.74,14.5,{size:21,color:C.muted});
  rect(0.65,8.17,14.7,0.51,C.pale,C.pale,0.06);
  text(d.takeaway,0.85,8.28,14.2,{size:16,bold:true,color:C.teal});
  text('HANDBOOK OF AUTHORIZATION · OUR MODEL · CORE SHAPES, NOT COMPLETE SCHEMAS',0.70,8.77,13.8,{size:9,color:C.muted});
  text(String(slideNo).padStart(2,'0'),14.72,8.65,0.6,{size:14,color:C.muted,align:'center'});
}
function render(d) {
  base(d);
  if(d.kind==='cover') {
    // Replace base with a dedicated cover for a clear opening hierarchy.
    sl._slideObjects.length=0;
    while(boxes.length && boxes[boxes.length-1].slide===slideNo)boxes.pop();
    svg=svg.slice(0,1);rect(0,0,W,H,C.ink,C.ink,0);rect(0.72,0.72,0.12,7.4,C.teal,C.teal,0);
    text('FOUNDATIONS / VOCABULARY / CANONICAL JSON',1.12,1.0,13,{size:16,bold:true,color:'83D4D6'});
    text(d.title,1.12,2.0,13.7,{size:56,bold:true,color:C.white});
    text(d.subtitle,1.15,4.4,12.6,{size:34,color:'83D4D6'});
    text('Definitions • field meanings • connected examples',1.15,5.30,13,{size:24,color:'D7E2ED'});
    text('Revision mechanics are intentionally outside this deck.\nRequired fields remain in the JSON; no contracts are changed.',1.15,6.5,12.8,{size:20,color:'BECFDF'});
    text('Based on the Handbook of Authorization · Working model',1.15,8.05,12.8,{size:14,color:'BECFDF'});
  } else if(d.kind==='split') {
    points(d.points);
    if(slideNo===19) {
      code(d.codes[0].value,'APPROVED GRANT-WIDE CONTROL',6.35,2.5,9,5.5);
    } else if(d.codes) {
      const firstH=d.codes[0].value.actor?2.55:1.90;
      const secondH=5.55-firstH-0.16;
      code(d.codes[0].value,d.codes[0].label,6.35,2.35,9,firstH);
      code(d.codes[1].value,d.codes[1].label,6.35,2.35+firstH+0.16,9,secondH);
    } else code(d.raw||d.code,d.codeLabel,6.35,2.5,9,5.5,!!d.raw);
  } else if(d.kind==='cards') {
    const six=d.cards.length===6;
    d.cards.forEach((p,i)=>box(p.term,p.text,0.72+(i%3)*5.0,2.7+(six?Math.floor(i/3)*2.48:0),4.72,six?2.15:4.52,{size:six?22:25}));
  } else if(d.kind==='gates') {
    box('REQUEST','Maya proposes an assignment',0.8,3.3,3.15,2.55,{size:25});
    box('AUTH ENDPOINT CHECK','May Maya perform this assignment operation for this recipient?',4.48,3.0,4.65,3.15,{size:24});
    box('AUTH BOUNDARY CHECK','Does the proposed authority fit valid sources and team ceilings?',9.7,3.0,5.35,3.15,{size:24});
    arrow(3.96,4.57,4.4,4.57);arrow(9.14,4.57,9.61,4.57);
    text('Only a request passing both checks may be persisted.',2.0,7.0,12,{size:26,bold:true,color:C.teal,align:'center'});
  } else if(d.kind==='lineage') {
    box('G1 · PARENT GRANT','Read + write\nFinance boundary',0.85,2.75,5.9,1.73,{size:20});
    box('TEAM1','Holds G1 through A1',9.0,2.75,5.9,1.73,{size:23});
    arrow(6.85,3.64,8.89,3.64);text('A1 assignment',6.97,3.12,1.88,{size:14,bold:true,color:C.teal,align:'center'});
    box('G2 · CHILD GRANT','Read only\nFinance AND C17',0.85,5.52,5.9,1.73,{size:20});
    box('TEAM2','Holds G2 through A2',9.0,5.52,5.9,1.73,{size:23});
    arrow(6.85,6.36,8.89,6.36);text('A2 assignment',6.97,5.87,1.88,{size:14,bold:true,color:C.teal,align:'center'});
    arrow(3.8,4.55,3.8,5.41);text('Parent grant ceiling',0.93,4.76,2.65,{size:16,color:C.teal});
    arrow(12,4.55,12,5.41);text('Parent team ceiling',12.25,4.76,2.65,{size:16,color:C.teal});
    text('Nutan → explicit Team2 membership → valid G2 authority',2.0,7.56,12,{size:22,bold:true,color:C.teal,align:'center'});
  } else if(d.kind==='flow') {
    d.steps.forEach((p,i)=>{box(p.term,p.text,0.72+i*3.79,3.15,3.42,3.32,{size:23});if(i<3)arrow(4.20+i*3.79,4.8,4.43+i*3.79,4.8);});
    text('Relationships and boundaries must remain connected throughout.',1.0,7.15,14,{size:23,color:C.muted,align:'center'});
  } else if(d.kind==='resolved') {
    const rows=[['Recipient route','Nutan → Team2 membership → A2 → G2'],['Required support','Team2 → Team1; A1 → G1 → valid G0 support'],['Permission','Certificate read only'],['Effective boundary','Finance AND C17, plus all inherited restrictions'],['Continuing limits','Live controls, validity and any delegation limits']];
    rows.forEach(([a,b],i)=>{rect(0.75,2.65+i*0.91,14.5,0.79,i%2?C.pale:C.white);text(a,1.0,2.85+i*0.91,3.35,{size:21,bold:true,color:C.teal});text(b,4.65,2.85+i*0.91,10.2,{size:21});});
    text('EXPLANATORY VIEW · RESOLVED-GRANT JSON IS STILL PENDING',0.95,7.56,14,{size:15,bold:true,color:C.amber});
  } else if(d.kind==='results') {
    text('DENY · evaluation completed',0.85,2.45,7.05,{size:23,bold:true,color:C.red});
    text('ERROR · evaluation could not finish',8.2,2.45,7.0,{size:23,bold:true,color:C.amber});
    code(deny,'APPROVED CORE SHAPE · ILLUSTRATIVE CODE',0.75,3.05,7.1,4.6);
    code(error,'APPROVED CORE SHAPE · ILLUSTRATIVE CODE',8.15,3.05,7.1,4.6);
  } else if(d.kind==='summary') {
    box('KEEP SEPARATE','Role ≠ grant ≠ assignment\nMembership ≠ ownership\nEnabled ≠ effective\nResolved ≠ allowed',0.8,2.7,7.0,4.48,{size:27});
    box('FORMATS STILL PENDING','Team / membership / ownership\nStandalone role and registration\nRoot and delegation evidence\nRequest / resolved-view envelopes',8.15,2.7,7.0,4.48,{size:23,accent:C.amber,fill:C.amberBg});
    text('Core record shapes are explained—not promoted into complete schemas.',1.0,7.5,14,{size:23,color:C.muted,align:'center'});
  }
  sl.addNotes(`${d.notes}\n\nSOURCE: handbook/slides/${d.source}\n\n${modelNotes}`);
  svg.push('</svg>');svgs.push(svg.join('\n'));
}

async function main() {
  for(let i=0;i<slides.length;i++){slideNo=i+1;render(slides[i]);}
  fs.mkdirSync(path.join(OUT,'preview'),{recursive:true});
  svgs.forEach((s,i)=>fs.writeFileSync(path.join(OUT,'preview',`slide-${String(i+1).padStart(2,'0')}.svg`),s));
  const page=`<!doctype html><html lang="en"><meta charset="utf-8"><title>Authorization — slide preview</title><style>body{margin:24px;background:#dce3eb;font:18px sans-serif;color:#14263d}h1{font-size:25px}.grid{display:grid;grid-template-columns:repeat(3,1fr);gap:18px}figure{margin:0}img{width:100%;display:block;border-radius:6px}figcaption{font-size:13px;padding:5px}a{color:inherit}@media print{.grid{display:block}figure{break-after:page}h1,p,figcaption{display:none}}</style><h1>Handbook of Authorization — Core Concepts &amp; JSON</h1><p>Layout preview generated from the same drawing primitives as the editable PowerPoint. Revision mechanics are outside scope.</p><div class="grid">${slides.map((d,i)=>`<figure><a href="preview/slide-${String(i+1).padStart(2,'0')}.svg"><img src="preview/slide-${String(i+1).padStart(2,'0')}.svg" alt="${esc(d.title)}"></a><figcaption>${i+1}. ${esc(d.title)}</figcaption></figure>`).join('')}</div></html>`;
  fs.writeFileSync(path.join(OUT,'preview.html'),page);
  fs.writeFileSync(path.join(OUT,'speaker-notes.md'),'# Core concepts & JSON — presenter notes\n\nRevision/adoption mechanics are intentionally outside the presentation.\n\n'+slides.map((d,i)=>`## ${i+1}. ${d.title.replace(/\n/g,' ')}\n\n${d.notes}\n\nSource: [handbook reference](${d.source}).\n`).join('\n'));
  fs.writeFileSync(path.join(OUT,'deck-manifest.json'),JSON.stringify({version:'1',artifact:'editorial slide manifest, not an Auth contract',slide_count:slides.length,examples,text_boxes:boxes},null,2));
  await pptx.writeFile({fileName:path.join(OUT,'authorization-core-concepts.pptx')});
  console.log(JSON.stringify({slides:slides.length,jsonExamples:examples.length,pptx:path.join(OUT,'authorization-core-concepts.pptx'),previews:svgs.length}));
}
main().catch(e=>{console.error(e);process.exitCode=1;});
