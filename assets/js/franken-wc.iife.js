var FRANKENWC=function(l){"use strict";var Mt;const W=globalThis,V=W.ShadowRoot&&(W.ShadyCSS===void 0||W.ShadyCSS.nativeShadow)&&"adoptedStyleSheets"in Document.prototype&&"replace"in CSSStyleSheet.prototype,ut=Symbol(),ct=new WeakMap;let jt=class{constructor(t,e,i){if(this._$cssResult$=!0,i!==ut)throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");this.cssText=t,this.t=e}get styleSheet(){let t=this.o;const e=this.t;if(V&&t===void 0){const i=e!==void 0&&e.length===1;i&&(t=ct.get(e)),t===void 0&&((this.o=t=new CSSStyleSheet).replaceSync(this.cssText),i&&ct.set(e,t))}return t}toString(){return this.cssText}};const xt=r=>new jt(typeof r=="string"?r:r+"",void 0,ut),zt=(r,t)=>{if(V)r.adoptedStyleSheets=t.map(e=>e instanceof CSSStyleSheet?e:e.styleSheet);else for(const e of t){const i=document.createElement("style"),s=W.litNonce;s!==void 0&&i.setAttribute("nonce",s),i.textContent=e.cssText,r.appendChild(i)}},dt=V?r=>r:r=>r instanceof CSSStyleSheet?(t=>{let e="";for(const i of t.cssRules)e+=i.cssText;return xt(e)})(r):r;const{is:Gt,defineProperty:qt,getOwnPropertyDescriptor:Yt,getOwnPropertyNames:Wt,getOwnPropertySymbols:Kt,getPrototypeOf:Jt}=Object,O=globalThis,pt=O.trustedTypes,Vt=pt?pt.emptyScript:"",F=O.reactiveElementPolyfillSupport,L=(r,t)=>r,K={toAttribute(r,t){switch(t){case Boolean:r=r?Vt:null;break;case Object:case Array:r=r==null?r:JSON.stringify(r)}return r},fromAttribute(r,t){let e=r;switch(t){case Boolean:e=r!==null;break;case Number:e=r===null?null:Number(r);break;case Object:case Array:try{e=JSON.parse(r)}catch{e=null}}return e}},Z=(r,t)=>!Gt(r,t),ft={attribute:!0,type:String,converter:K,reflect:!1,hasChanged:Z};Symbol.metadata??(Symbol.metadata=Symbol("metadata")),O.litPropertyMetadata??(O.litPropertyMetadata=new WeakMap);class N extends HTMLElement{static addInitializer(t){this._$Ei(),(this.l??(this.l=[])).push(t)}static get observedAttributes(){return this.finalize(),this._$Eh&&[...this._$Eh.keys()]}static createProperty(t,e=ft){if(e.state&&(e.attribute=!1),this._$Ei(),this.elementProperties.set(t,e),!e.noAccessor){const i=Symbol(),s=this.getPropertyDescriptor(t,i,e);s!==void 0&&qt(this.prototype,t,s)}}static getPropertyDescriptor(t,e,i){const{get:s,set:o}=Yt(this.prototype,t)??{get(){return this[e]},set(n){this[e]=n}};return{get(){return s==null?void 0:s.call(this)},set(n){const h=s==null?void 0:s.call(this);o.call(this,n),this.requestUpdate(t,h,i)},configurable:!0,enumerable:!0}}static getPropertyOptions(t){return this.elementProperties.get(t)??ft}static _$Ei(){if(this.hasOwnProperty(L("elementProperties")))return;const t=Jt(this);t.finalize(),t.l!==void 0&&(this.l=[...t.l]),this.elementProperties=new Map(t.elementProperties)}static finalize(){if(this.hasOwnProperty(L("finalized")))return;if(this.finalized=!0,this._$Ei(),this.hasOwnProperty(L("properties"))){const e=this.properties,i=[...Wt(e),...Kt(e)];for(const s of i)this.createProperty(s,e[s])}const t=this[Symbol.metadata];if(t!==null){const e=litPropertyMetadata.get(t);if(e!==void 0)for(const[i,s]of e)this.elementProperties.set(i,s)}this._$Eh=new Map;for(const[e,i]of this.elementProperties){const s=this._$Eu(e,i);s!==void 0&&this._$Eh.set(s,e)}this.elementStyles=this.finalizeStyles(this.styles)}static finalizeStyles(t){const e=[];if(Array.isArray(t)){const i=new Set(t.flat(1/0).reverse());for(const s of i)e.unshift(dt(s))}else t!==void 0&&e.push(dt(t));return e}static _$Eu(t,e){const i=e.attribute;return i===!1?void 0:typeof i=="string"?i:typeof t=="string"?t.toLowerCase():void 0}constructor(){super(),this._$Ep=void 0,this.isUpdatePending=!1,this.hasUpdated=!1,this._$Em=null,this._$Ev()}_$Ev(){var t;this._$ES=new Promise(e=>this.enableUpdating=e),this._$AL=new Map,this._$E_(),this.requestUpdate(),(t=this.constructor.l)==null||t.forEach(e=>e(this))}addController(t){var e;(this._$EO??(this._$EO=new Set)).add(t),this.renderRoot!==void 0&&this.isConnected&&((e=t.hostConnected)==null||e.call(t))}removeController(t){var e;(e=this._$EO)==null||e.delete(t)}_$E_(){const t=new Map,e=this.constructor.elementProperties;for(const i of e.keys())this.hasOwnProperty(i)&&(t.set(i,this[i]),delete this[i]);t.size>0&&(this._$Ep=t)}createRenderRoot(){const t=this.shadowRoot??this.attachShadow(this.constructor.shadowRootOptions);return zt(t,this.constructor.elementStyles),t}connectedCallback(){var t;this.renderRoot??(this.renderRoot=this.createRenderRoot()),this.enableUpdating(!0),(t=this._$EO)==null||t.forEach(e=>{var i;return(i=e.hostConnected)==null?void 0:i.call(e)})}enableUpdating(t){}disconnectedCallback(){var t;(t=this._$EO)==null||t.forEach(e=>{var i;return(i=e.hostDisconnected)==null?void 0:i.call(e)})}attributeChangedCallback(t,e,i){this._$AK(t,i)}_$EC(t,e){var o;const i=this.constructor.elementProperties.get(t),s=this.constructor._$Eu(t,i);if(s!==void 0&&i.reflect===!0){const n=(((o=i.converter)==null?void 0:o.toAttribute)!==void 0?i.converter:K).toAttribute(e,i.type);this._$Em=t,n==null?this.removeAttribute(s):this.setAttribute(s,n),this._$Em=null}}_$AK(t,e){var o;const i=this.constructor,s=i._$Eh.get(t);if(s!==void 0&&this._$Em!==s){const n=i.getPropertyOptions(s),h=typeof n.converter=="function"?{fromAttribute:n.converter}:((o=n.converter)==null?void 0:o.fromAttribute)!==void 0?n.converter:K;this._$Em=s,this[s]=h.fromAttribute(e,n.type),this._$Em=null}}requestUpdate(t,e,i){if(t!==void 0){if(i??(i=this.constructor.getPropertyOptions(t)),!(i.hasChanged??Z)(this[t],e))return;this.P(t,e,i)}this.isUpdatePending===!1&&(this._$ES=this._$ET())}P(t,e,i){this._$AL.has(t)||this._$AL.set(t,e),i.reflect===!0&&this._$Em!==t&&(this._$Ej??(this._$Ej=new Set)).add(t)}async _$ET(){this.isUpdatePending=!0;try{await this._$ES}catch(e){Promise.reject(e)}const t=this.scheduleUpdate();return t!=null&&await t,!this.isUpdatePending}scheduleUpdate(){return this.performUpdate()}performUpdate(){var i;if(!this.isUpdatePending)return;if(!this.hasUpdated){if(this.renderRoot??(this.renderRoot=this.createRenderRoot()),this._$Ep){for(const[o,n]of this._$Ep)this[o]=n;this._$Ep=void 0}const s=this.constructor.elementProperties;if(s.size>0)for(const[o,n]of s)n.wrapped!==!0||this._$AL.has(o)||this[o]===void 0||this.P(o,this[o],n)}let t=!1;const e=this._$AL;try{t=this.shouldUpdate(e),t?(this.willUpdate(e),(i=this._$EO)==null||i.forEach(s=>{var o;return(o=s.hostUpdate)==null?void 0:o.call(s)}),this.update(e)):this._$EU()}catch(s){throw t=!1,this._$EU(),s}t&&this._$AE(e)}willUpdate(t){}_$AE(t){var e;(e=this._$EO)==null||e.forEach(i=>{var s;return(s=i.hostUpdated)==null?void 0:s.call(i)}),this.hasUpdated||(this.hasUpdated=!0,this.firstUpdated(t)),this.updated(t)}_$EU(){this._$AL=new Map,this.isUpdatePending=!1}get updateComplete(){return this.getUpdateComplete()}getUpdateComplete(){return this._$ES}shouldUpdate(t){return!0}update(t){this._$Ej&&(this._$Ej=this._$Ej.forEach(e=>this._$EC(e,this[e]))),this._$EU()}updated(t){}firstUpdated(t){}}N.elementStyles=[],N.shadowRootOptions={mode:"open"},N[L("elementProperties")]=new Map,N[L("finalized")]=new Map,F==null||F({ReactiveElement:N}),(O.reactiveElementVersions??(O.reactiveElementVersions=[])).push("2.0.4");const M=globalThis,J=M.trustedTypes,$t=J?J.createPolicy("lit-html",{createHTML:r=>r}):void 0,Q="$lit$",_=`lit$${Math.random().toFixed(9).slice(2)}$`,X="?"+_,Ft=`<${X}>`,T=document,B=()=>T.createComment(""),j=r=>r===null||typeof r!="object"&&typeof r!="function",gt=Array.isArray,mt=r=>gt(r)||typeof(r==null?void 0:r[Symbol.iterator])=="function",tt=`[ 	
\f\r]`,x=/<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g,vt=/-->/g,yt=/>/g,I=RegExp(`>|${tt}(?:([^\\s"'>=/]+)(${tt}*=${tt}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`,"g"),bt=/'/g,wt=/"/g,At=/^(?:script|style|textarea|title)$/i,Zt=r=>(t,...e)=>({_$litType$:r,strings:t,values:e}),g=Zt(1),C=Symbol.for("lit-noChange"),m=Symbol.for("lit-nothing"),_t=new WeakMap,U=T.createTreeWalker(T,129);function kt(r,t){if(!Array.isArray(r)||!r.hasOwnProperty("raw"))throw Error("invalid template strings array");return $t!==void 0?$t.createHTML(t):t}const Et=(r,t)=>{const e=r.length-1,i=[];let s,o=t===2?"<svg>":"",n=x;for(let h=0;h<e;h++){const a=r[h];let p,f,u=-1,$=0;for(;$<a.length&&(n.lastIndex=$,f=n.exec(a),f!==null);)$=n.lastIndex,n===x?f[1]==="!--"?n=vt:f[1]!==void 0?n=yt:f[2]!==void 0?(At.test(f[2])&&(s=RegExp("</"+f[2],"g")),n=I):f[3]!==void 0&&(n=I):n===I?f[0]===">"?(n=s??x,u=-1):f[1]===void 0?u=-2:(u=n.lastIndex-f[2].length,p=f[1],n=f[3]===void 0?I:f[3]==='"'?wt:bt):n===wt||n===bt?n=I:n===vt||n===yt?n=x:(n=I,s=void 0);const d=n===I&&r[h+1].startsWith("/>")?" ":"";o+=n===x?a+Ft:u>=0?(i.push(p),a.slice(0,u)+Q+a.slice(u)+_+d):a+_+(u===-2?h:d)}return[kt(r,o+(r[e]||"<?>")+(t===2?"</svg>":"")),i]};class z{constructor({strings:t,_$litType$:e},i){let s;this.parts=[];let o=0,n=0;const h=t.length-1,a=this.parts,[p,f]=Et(t,e);if(this.el=z.createElement(p,i),U.currentNode=this.el.content,e===2){const u=this.el.content.firstChild;u.replaceWith(...u.childNodes)}for(;(s=U.nextNode())!==null&&a.length<h;){if(s.nodeType===1){if(s.hasAttributes())for(const u of s.getAttributeNames())if(u.endsWith(Q)){const $=f[n++],d=s.getAttribute(u).split(_),y=/([.?@])?(.*)/.exec($);a.push({type:1,index:o,name:y[2],strings:d,ctor:y[1]==="."?Ot:y[1]==="?"?Ct:y[1]==="@"?Pt:G}),s.removeAttribute(u)}else u.startsWith(_)&&(a.push({type:6,index:o}),s.removeAttribute(u));if(At.test(s.tagName)){const u=s.textContent.split(_),$=u.length-1;if($>0){s.textContent=J?J.emptyScript:"";for(let d=0;d<$;d++)s.append(u[d],B()),U.nextNode(),a.push({type:2,index:++o});s.append(u[$],B())}}}else if(s.nodeType===8)if(s.data===X)a.push({type:2,index:o});else{let u=-1;for(;(u=s.data.indexOf(_,u+1))!==-1;)a.push({type:7,index:o}),u+=_.length-1}o++}}static createElement(t,e){const i=T.createElement("template");return i.innerHTML=t,i}}function R(r,t,e=r,i){var n,h;if(t===C)return t;let s=i!==void 0?(n=e._$Co)==null?void 0:n[i]:e._$Cl;const o=j(t)?void 0:t._$litDirective$;return(s==null?void 0:s.constructor)!==o&&((h=s==null?void 0:s._$AO)==null||h.call(s,!1),o===void 0?s=void 0:(s=new o(r),s._$AT(r,e,i)),i!==void 0?(e._$Co??(e._$Co=[]))[i]=s:e._$Cl=s),s!==void 0&&(t=R(r,s._$AS(r,t.values),s,i)),t}class St{constructor(t,e){this._$AV=[],this._$AN=void 0,this._$AD=t,this._$AM=e}get parentNode(){return this._$AM.parentNode}get _$AU(){return this._$AM._$AU}u(t){const{el:{content:e},parts:i}=this._$AD,s=((t==null?void 0:t.creationScope)??T).importNode(e,!0);U.currentNode=s;let o=U.nextNode(),n=0,h=0,a=i[0];for(;a!==void 0;){if(n===a.index){let p;a.type===2?p=new H(o,o.nextSibling,this,t):a.type===1?p=new a.ctor(o,a.name,a.strings,this,t):a.type===6&&(p=new Tt(o,this,t)),this._$AV.push(p),a=i[++h]}n!==(a==null?void 0:a.index)&&(o=U.nextNode(),n++)}return U.currentNode=T,s}p(t){let e=0;for(const i of this._$AV)i!==void 0&&(i.strings!==void 0?(i._$AI(t,i,e),e+=i.strings.length-2):i._$AI(t[e])),e++}}class H{get _$AU(){var t;return((t=this._$AM)==null?void 0:t._$AU)??this._$Cv}constructor(t,e,i,s){this.type=2,this._$AH=m,this._$AN=void 0,this._$AA=t,this._$AB=e,this._$AM=i,this.options=s,this._$Cv=(s==null?void 0:s.isConnected)??!0}get parentNode(){let t=this._$AA.parentNode;const e=this._$AM;return e!==void 0&&(t==null?void 0:t.nodeType)===11&&(t=e.parentNode),t}get startNode(){return this._$AA}get endNode(){return this._$AB}_$AI(t,e=this){t=R(this,t,e),j(t)?t===m||t==null||t===""?(this._$AH!==m&&this._$AR(),this._$AH=m):t!==this._$AH&&t!==C&&this._(t):t._$litType$!==void 0?this.$(t):t.nodeType!==void 0?this.T(t):mt(t)?this.k(t):this._(t)}S(t){return this._$AA.parentNode.insertBefore(t,this._$AB)}T(t){this._$AH!==t&&(this._$AR(),this._$AH=this.S(t))}_(t){this._$AH!==m&&j(this._$AH)?this._$AA.nextSibling.data=t:this.T(T.createTextNode(t)),this._$AH=t}$(t){var o;const{values:e,_$litType$:i}=t,s=typeof i=="number"?this._$AC(t):(i.el===void 0&&(i.el=z.createElement(kt(i.h,i.h[0]),this.options)),i);if(((o=this._$AH)==null?void 0:o._$AD)===s)this._$AH.p(e);else{const n=new St(s,this),h=n.u(this.options);n.p(e),this.T(h),this._$AH=n}}_$AC(t){let e=_t.get(t.strings);return e===void 0&&_t.set(t.strings,e=new z(t)),e}k(t){gt(this._$AH)||(this._$AH=[],this._$AR());const e=this._$AH;let i,s=0;for(const o of t)s===e.length?e.push(i=new H(this.S(B()),this.S(B()),this,this.options)):i=e[s],i._$AI(o),s++;s<e.length&&(this._$AR(i&&i._$AB.nextSibling,s),e.length=s)}_$AR(t=this._$AA.nextSibling,e){var i;for((i=this._$AP)==null?void 0:i.call(this,!1,!0,e);t&&t!==this._$AB;){const s=t.nextSibling;t.remove(),t=s}}setConnected(t){var e;this._$AM===void 0&&(this._$Cv=t,(e=this._$AP)==null||e.call(this,t))}}class G{get tagName(){return this.element.tagName}get _$AU(){return this._$AM._$AU}constructor(t,e,i,s,o){this.type=1,this._$AH=m,this._$AN=void 0,this.element=t,this.name=e,this._$AM=s,this.options=o,i.length>2||i[0]!==""||i[1]!==""?(this._$AH=Array(i.length-1).fill(new String),this.strings=i):this._$AH=m}_$AI(t,e=this,i,s){const o=this.strings;let n=!1;if(o===void 0)t=R(this,t,e,0),n=!j(t)||t!==this._$AH&&t!==C,n&&(this._$AH=t);else{const h=t;let a,p;for(t=o[0],a=0;a<o.length-1;a++)p=R(this,h[i+a],e,a),p===C&&(p=this._$AH[a]),n||(n=!j(p)||p!==this._$AH[a]),p===m?t=m:t!==m&&(t+=(p??"")+o[a+1]),this._$AH[a]=p}n&&!s&&this.j(t)}j(t){t===m?this.element.removeAttribute(this.name):this.element.setAttribute(this.name,t??"")}}class Ot extends G{constructor(){super(...arguments),this.type=3}j(t){this.element[this.name]=t===m?void 0:t}}class Ct extends G{constructor(){super(...arguments),this.type=4}j(t){this.element.toggleAttribute(this.name,!!t&&t!==m)}}class Pt extends G{constructor(t,e,i,s,o){super(t,e,i,s,o),this.type=5}_$AI(t,e=this){if((t=R(this,t,e,0)??m)===C)return;const i=this._$AH,s=t===m&&i!==m||t.capture!==i.capture||t.once!==i.once||t.passive!==i.passive,o=t!==m&&(i===m||s);s&&this.element.removeEventListener(this.name,this,i),o&&this.element.addEventListener(this.name,this,t),this._$AH=t}handleEvent(t){var e;typeof this._$AH=="function"?this._$AH.call(((e=this.options)==null?void 0:e.host)??this.element,t):this._$AH.handleEvent(t)}}class Tt{constructor(t,e,i){this.element=t,this.type=6,this._$AN=void 0,this._$AM=e,this.options=i}get _$AU(){return this._$AM._$AU}_$AI(t){R(this,t)}}const Qt={P:Q,A:_,C:X,M:1,L:Et,R:St,D:mt,V:R,I:H,H:G,N:Ct,U:Pt,B:Ot,F:Tt},et=M.litHtmlPolyfillSupport;et==null||et(z,H),(M.litHtmlVersions??(M.litHtmlVersions=[])).push("3.1.4");const Xt=(r,t,e)=>{const i=(e==null?void 0:e.renderBefore)??t;let s=i._$litPart$;if(s===void 0){const o=(e==null?void 0:e.renderBefore)??null;i._$litPart$=s=new H(t.insertBefore(B(),o),o,void 0,e??{})}return s._$AI(r),s};let k=class extends N{constructor(){super(...arguments),this.renderOptions={host:this},this._$Do=void 0}createRenderRoot(){var e;const t=super.createRenderRoot();return(e=this.renderOptions).renderBefore??(e.renderBefore=t.firstChild),t}update(t){const e=this.render();this.hasUpdated||(this.renderOptions.isConnected=this.isConnected),super.update(t),this._$Do=Xt(e,this.renderRoot,this.renderOptions)}connectedCallback(){var t;super.connectedCallback(),(t=this._$Do)==null||t.setConnected(!0)}disconnectedCallback(){var t;super.disconnectedCallback(),(t=this._$Do)==null||t.setConnected(!1)}render(){return C}};k._$litElement$=!0,k.finalized=!0,(Mt=globalThis.litElementHydrateSupport)==null||Mt.call(globalThis,{LitElement:k});const it=globalThis.litElementPolyfillSupport;it==null||it({LitElement:k}),(globalThis.litElementVersions??(globalThis.litElementVersions=[])).push("4.0.6");const q=r=>(t,e)=>{e!==void 0?e.addInitializer(()=>{customElements.define(r,t)}):customElements.define(r,t)};const te={attribute:!0,type:String,converter:K,reflect:!1,hasChanged:Z},ee=(r=te,t,e)=>{const{kind:i,metadata:s}=e;let o=globalThis.litPropertyMetadata.get(s);if(o===void 0&&globalThis.litPropertyMetadata.set(s,o=new Map),o.set(e.name,r),i==="accessor"){const{name:n}=e;return{set(h){const a=t.get.call(this);t.set.call(this,h),this.requestUpdate(n,a,r)},init(h){return h!==void 0&&this.P(n,void 0,r),h}}}if(i==="setter"){const{name:n}=e;return function(h){const a=this[n];t.call(this,h),this.requestUpdate(n,a,r)}}throw Error("Unsupported decorator location: "+i)};function c(r){return(t,e)=>typeof e=="object"?ee(r,t,e):((i,s,o)=>{const n=s.hasOwnProperty(o);return s.constructor.createProperty(o,n?{...i,wrapped:!0}:i),n?Object.getOwnPropertyDescriptor(s,o):void 0})(r,t,e)}function v(r){return c({...r,state:!0,attribute:!1})}const It={ATTRIBUTE:1,CHILD:2,PROPERTY:3,BOOLEAN_ATTRIBUTE:4,EVENT:5,ELEMENT:6},Ut=r=>(...t)=>({_$litDirective$:r,values:t});class Rt{constructor(t){}get _$AU(){return this._$AM._$AU}_$AT(t,e,i){this._$Ct=t,this._$AM=e,this._$Ci=i}_$AS(t,e){return this.update(t,e)}update(t,e){return this.render(...e)}}const{I:ie}=Qt,Dt=()=>document.createComment(""),Y=(r,t,e)=>{var o;const i=r._$AA.parentNode,s=t===void 0?r._$AB:t._$AA;if(e===void 0){const n=i.insertBefore(Dt(),s),h=i.insertBefore(Dt(),s);e=new ie(n,h,r,r.options)}else{const n=e._$AB.nextSibling,h=e._$AM,a=h!==r;if(a){let p;(o=e._$AQ)==null||o.call(e,r),e._$AM=r,e._$AP!==void 0&&(p=r._$AU)!==h._$AU&&e._$AP(p)}if(n!==s||a){let p=e._$AA;for(;p!==n;){const f=p.nextSibling;i.insertBefore(p,s),p=f}}}return e},D=(r,t,e=r)=>(r._$AI(t,e),r),se={},re=(r,t=se)=>r._$AH=t,oe=r=>r._$AH,st=r=>{var i;(i=r._$AP)==null||i.call(r,!1,!0);let t=r._$AA;const e=r._$AB.nextSibling;for(;t!==e;){const s=t.nextSibling;t.remove(),t=s}};const Nt=(r,t,e)=>{const i=new Map;for(let s=t;s<=e;s++)i.set(r[s],s);return i},rt=Ut(class extends Rt{constructor(r){if(super(r),r.type!==It.CHILD)throw Error("repeat() can only be used in text expressions")}dt(r,t,e){let i;e===void 0?e=t:t!==void 0&&(i=t);const s=[],o=[];let n=0;for(const h of r)s[n]=i?i(h,n):n,o[n]=e(h,n),n++;return{values:o,keys:s}}render(r,t,e){return this.dt(r,t,e).values}update(r,[t,e,i]){const s=oe(r),{values:o,keys:n}=this.dt(t,e,i);if(!Array.isArray(s))return this.ut=n,o;const h=this.ut??(this.ut=[]),a=[];let p,f,u=0,$=s.length-1,d=0,y=o.length-1;for(;u<=$&&d<=y;)if(s[u]===null)u++;else if(s[$]===null)$--;else if(h[u]===n[d])a[d]=D(s[u],o[d]),u++,d++;else if(h[$]===n[y])a[y]=D(s[$],o[y]),$--,y--;else if(h[u]===n[y])a[y]=D(s[u],o[y]),Y(r,a[y+1],s[u]),u++,y--;else if(h[$]===n[d])a[d]=D(s[$],o[d]),Y(r,s[u],s[$]),$--,d++;else if(p===void 0&&(p=Nt(n,d,y),f=Nt(h,u,$)),p.has(h[u]))if(p.has(h[$])){const S=f.get(n[d]),ht=S!==void 0?s[S]:null;if(ht===null){const Bt=Y(r,s[u]);D(Bt,o[d]),a[d]=Bt}else a[d]=D(ht,o[d]),Y(r,s[u],ht),s[S]=null;d++}else st(s[$]),$--;else st(s[u]),u++;for(;d<=y;){const S=Y(r,a[y+1]);D(S,o[d]),a[d++]=S}for(;u<=$;){const S=s[u++];S!==null&&st(S)}return this.ut=n,re(r,a),C}});function ot(r){try{if(r.startsWith("{"))return JSON.parse(r);const t={};return r.replace(/[;\s]+$/,"").split(";").forEach(e=>{const i=e.trim().split(/:(.*)/);t[i[0].trim()]=i[1].trim()}),t}catch{return{}}}function nt(r){if(/^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$/.test(r))return r}function lt(r){if(/^(\d*\.?\d+)(px|cm|mm|in|pt|pc|em|ex|ch|rem|vw|vh|vmin|vmax|%)$/.test(r))return r}function ne(r){if(["none","hidden","dotted","dashed","solid","double","groove","ridge","inset","outset"].includes(r))return r}function le(r){if(/^(\d*\.?\d+)(ms|s)$/.test(r))return r}function ae(r=5){const t="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";return Array.from({length:r},()=>t.charAt(Math.floor(Math.random()*t.length))).join("")}var he=Object.defineProperty,ue=Object.getOwnPropertyDescriptor,b=(r,t,e,i)=>{for(var s=i>1?void 0:i?ue(t,e):t,o=r.length-1,n;o>=0;o--)(n=r[o])&&(s=(i?n(t,e,s):n(s))||s);return i&&s&&he(t,e,s),s};l.Select=class extends k{constructor(){super(...arguments),this.name="",this.multiple=!1,this.disabled=!1,this.placeholder="",this.searchable=!1,this.error=!1,this.i18n="",this.$term="",this.$options=[],this.$filteredOptions=this.$options,this.$focused=-1,this.$selected=[],this.$isOpen=!1,this.$i18n={"selection-count-text":":n: options selected"}}navigate(t){const e=o=>o.type!=="label"&&o.disabled!==!0;let i=this.$focused;const s=t==="up"?-1:1;do if(i+=s,i<0){for(i=this.$filteredOptions.length-1;i>=0&&!e(this.$filteredOptions[i]);)i--;break}else if(i>=this.$filteredOptions.length){for(i=0;i<this.$filteredOptions.length&&!e(this.$filteredOptions[i]);)i++;break}while(!e(this.$filteredOptions[i]));return i}addOption(t,e){t.selected===!0&&(this.multiple===!1?this.$selected=[t.value]:this.$selected.push(t.value));let i;t.hasAttribute("value")?i=t.getAttribute("value")||"":i=t.textContent||"",this.$options.push({type:"option",value:i,text:t.textContent||"",disabled:e===!0?!0:t.disabled,selected:t.selected})}connectedCallback(){super.connectedCallback(),Array.from(this.children).map(t=>{if(t.nodeName==="OPTGROUP"){const e=t;this.$options.push({type:"label",text:e.getAttribute("label")||""}),Array.from(e.children).map(i=>{const s=i;this.addOption(s,e.disabled)})}if(t.nodeName==="OPTION"){const e=t;this.addOption(e)}}),this.multiple===!1&&this.$selected.length===1&&(this.$focused=this.$options.findIndex(t=>t.value===this.$selected[0])),this.i18n&&(this.$i18n=ot(this.i18n)),document.addEventListener("click",this.onClickAway.bind(this)),this.innerHTML="",this.removeAttribute("uk-cloak")}disconnectedCallback(){super.disconnectedCallback(),document.removeEventListener("click",this.onClickAway)}createRenderRoot(){return this}updated(t){var e;if(t.has("$focused")&&this.$isOpen===!0&&this.focusLi(),t.has("$isOpen"))if(this.$isOpen===!0){this.focusLi(!1);const i=window.innerHeight,s=this.renderRoot.querySelector("div.uk-dropdown"),o=this.renderRoot.querySelector("button"),n={dropdown:s.getBoundingClientRect(),button:o.getBoundingClientRect()};n.button.bottom+((e=n.dropdown)==null?void 0:e.height)>i&&(s.style.bottom=`${n.button.height+4}px`),this.dispatchEvent(new CustomEvent("uk-select:shown",{detail:{value:!0},bubbles:!0,composed:!0}))}else this.updateComplete.then(()=>{this.$term="",this.multiple===!1?this.$focused=this.$options.findIndex(i=>i.value===this.$selected[0]):this.$focused=-1}),this.dispatchEvent(new CustomEvent("uk-select:hidden",{detail:{value:!0},bubbles:!0,composed:!0}));t.has("$term")&&this.updateComplete.then(()=>{if(this.$term===""){this.$filteredOptions=this.$options;return}this.$filteredOptions=this.$options.filter(i=>{var s;return(s=i.value)==null?void 0:s.toLowerCase().includes(this.$term.toLowerCase())})})}render(){return g`
      <div class="uk-custom-select">
        <button
          class="uk-fake-input uk-flex uk-flex-between ${this.error===!0?"uk-form-danger":""}"
          type="button"
          .disabled=${this.disabled}
          @click="${this.toggle}"
          @keydown=${t=>{if(this.$isOpen===!0)switch(t.key){case"Escape":this.$isOpen=!1;break;case"ArrowDown":t.preventDefault(),this.$focused=this.navigate("down");break;case"ArrowUp":t.preventDefault(),this.$focused=this.navigate("up");break;case"Enter":t.preventDefault(),this.select(this.$focused);break;case" ":t.preventDefault(),this.select(this.$focused);break;case"Tab":this.searchable===!1&&(this.$isOpen=!1);break}else switch(t.key){case"ArrowDown":t.preventDefault(),this.$focused=this.navigate("down"),this.$isOpen=!0;break;case"ArrowUp":t.preventDefault(),this.$focused=this.navigate("up"),this.$isOpen=!0;break}}}
        >
          <span> ${this.text()} </span>
          <svg
            class="opacity-50"
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="lucide lucide-chevrons-up-down"
          >
            <path d="m7 15 5 5 5-5" />
            <path d="m7 9 5-5 5 5" />
          </svg>
        </button>
        ${this.$isOpen===!0?g`
              <div class="uk-drop uk-dropdown uk-open" tabindex="-1">
                ${this.searchable===!0?g`
                      <div class="uk-custom-select-search">
                        <svg
                          xmlns="http://www.w3.org/2000/svg"
                          width="16"
                          height="16"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          stroke-width="2"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                          class="lucide lucide-search"
                        >
                          <circle cx="11" cy="11" r="8" />
                          <path d="m21 21-4.3-4.3" />
                        </svg>
                        <input
                          placeholder="Search"
                          type="text"
                          .value="${this.$term}"
                          @keydown=${t=>{var e;if(this.$isOpen===!0)switch(t.key){case"Escape":this.$isOpen=!1,(e=this.renderRoot.querySelector("button"))==null||e.focus();break;case"ArrowDown":t.preventDefault(),this.$focused=this.navigate("down");break;case"ArrowUp":t.preventDefault(),this.$focused=this.navigate("up");break;case"Enter":t.preventDefault(),this.select(this.$focused);break;case"Tab":!t.altKey&&!t.shiftKey&&!t.ctrlKey&&!t.metaKey&&(this.$isOpen=!1);break}}}
                          @input=${t=>{const e=t.target;this.$term=e.value}}
                        />
                      </div>
                    `:""}
                ${this.$filteredOptions.length>0?g`
                      <hr class="uk-hr" />
                      <ul class="uk-dropdown-nav" tabindex="-1">
                        ${rt(this.$filteredOptions,t=>t.value,(t,e)=>g`${t.type==="label"?g`<li class="uk-nav-header">
                                  ${t.text}
                                </li>`:g`<li
                                  class="${t.disabled===!0?"uk-disabled opacity-50":""} ${this.$focused===e?"uk-active":""}"
                                  tabindex="-1"
                                  @click=${()=>this.select(e)}
                                >
                                  <a tabindex="-1">
                                    <span>${t.text}</span>
                                    ${this.$selected.includes(t.value)?g`<svg
                                          xmlns="http://www.w3.org/2000/svg"
                                          width="16"
                                          height="16"
                                          viewBox="0 0 24 24"
                                          fill="none"
                                          stroke="currentColor"
                                          stroke-width="2"
                                          stroke-linecap="round"
                                          stroke-linejoin="round"
                                          class="lucide lucide-check"
                                        >
                                          <path d="M20 6 9 17l-5-5" />
                                        </svg>`:""}
                                  </a>
                                </li>`}`)}
                      </ul>
                    `:""}
              </div>
            `:""}
        ${this.name&&this.$selected.length>0?g`${this.multiple===!1?g`
                  <input
                    name="${this.name}"
                    type="hidden"
                    value="${this.$selected[0]}"
                  />
                `:this.$selected.map(t=>g`
                    <input name="${this.name}[]" type="hidden" value="${t}" />
                  `)}`:""}
      </div>
    `}text(){var t,e;return this.$selected.length===0?this.placeholder!==""?this.placeholder:"Select an option":this.multiple===!1?(t=this.$options.find(i=>i.value===this.$selected[0]))==null?void 0:t.text:this.$selected.length===1?(e=this.$options.find(i=>i.value===this.$selected[0]))==null?void 0:e.text:this.$i18n["selection-count-text"].replace(":n:",this.$selected.length.toString())}toggle(){this.$options.length!==0&&(this.$isOpen=!this.$isOpen)}focusLi(t=!0){const e=this.renderRoot.querySelector("ul");if(e){const i=e.querySelectorAll("li");if(this.$focused>=0&&this.$focused<i.length){const s=i[this.$focused],o={ul:e.getBoundingClientRect(),li:s.getBoundingClientRect()},n=s.offsetTop-e.offsetTop-o.ul.height/2+o.li.height/2;t===!0?e.scrollTo({top:n,behavior:"smooth"}):e.scrollTop=n}}}select(t){var i,s;if(t===-1){this.$isOpen=!1,(i=this.renderRoot.querySelector("button"))==null||i.focus();return}let e=null;t!==void 0&&(e=this.$filteredOptions[t]),!(e&&(e.type==="label"||e.disabled===!0))&&(this.multiple===!1?(t!==void 0&&(this.$focused=t,this.$selected=[e==null?void 0:e.value]),this.$isOpen=!1,(s=this.renderRoot.querySelector("button"))==null||s.focus(),this.dispatchEvent(new CustomEvent("uk-select:input",{detail:{value:this.$selected[0]},bubbles:!0,composed:!0}))):(t!==void 0&&(this.$selected.findIndex(o=>o===(e==null?void 0:e.value))===-1?this.$selected.push(e==null?void 0:e.value):this.$selected=this.$selected.filter(o=>o!==(e==null?void 0:e.value)),this.requestUpdate()),this.dispatchEvent(new CustomEvent("uk-select:input",{detail:{value:this.$selected},bubbles:!0,composed:!0}))))}onClickAway(t){this.$isOpen&&!this.renderRoot.contains(t.target)&&(this.$isOpen=!1)}},b([c({type:String})],l.Select.prototype,"name",2),b([c({type:Boolean})],l.Select.prototype,"multiple",2),b([c({type:Boolean})],l.Select.prototype,"disabled",2),b([c({type:String})],l.Select.prototype,"placeholder",2),b([c({type:Boolean})],l.Select.prototype,"searchable",2),b([c({type:Boolean})],l.Select.prototype,"error",2),b([c({type:String})],l.Select.prototype,"i18n",2),b([v()],l.Select.prototype,"$term",2),b([v()],l.Select.prototype,"$options",2),b([v()],l.Select.prototype,"$filteredOptions",2),b([v()],l.Select.prototype,"$focused",2),b([v()],l.Select.prototype,"$selected",2),b([v()],l.Select.prototype,"$isOpen",2),b([v()],l.Select.prototype,"$i18n",2),l.Select=b([q("uk-select")],l.Select);class ce{constructor(){this.subscribers=[],this.observer=new MutationObserver(()=>{const t=document.documentElement.classList.contains("dark");this.subscribers.forEach(e=>e(t))}),this.observer.observe(document.documentElement,{attributeFilter:["class"],attributeOldValue:!0})}subscribe(t){this.subscribers.push(t)}unsubscribe(t){const e=this.subscribers.indexOf(t);e!==-1&&this.subscribers.splice(e,1)}}const Ht=new ce;var de=Object.defineProperty,pe=Object.getOwnPropertyDescriptor,w=(r,t,e,i)=>{for(var s=i>1?void 0:i?pe(t,e):t,o=r.length-1,n;o>=0;o--)(n=r[o])&&(s=(i?n(t,e,s):n(s))||s);return i&&s&&de(t,e,s),s};l.Glow=class extends k{constructor(){super(...arguments),this.$mode="light",this.$defaults={},this.$shouldRender=!1,this["border-radius"]="0.6rem",this["border-width"]="0px",this["border-style"]="solid",this["border-color"]="#1e293b",this.width="auto",this.height="auto",this["background-color"]="#09090b",this["glow-color"]="#06b6d4",this["glow-width"]="0.125rem",this["glow-speed"]="10s",this.color="#fafafa"}connectedCallback(){super.connectedCallback(),this.initializeDefaults(),this.$mode=document.documentElement.classList.contains("dark")?"dark":"light",Ht.subscribe(t=>{this.$mode=t===!0?"dark":"light"}),this.removeAttribute("uk-cloak")}disconnectedCallback(){Ht.unsubscribe(()=>{})}initializeDefaults(){try{this.$defaults.colors=this.validateColors()}catch(t){return console.error(t)}try{this.$defaults.border=this.validateBorder()}catch(t){return console.error(t)}try{const t=this.validateSize();this.$defaults.width=t.width,this.$defaults.height=t.height}catch(t){return console.error(t)}try{const t=this.validateGlow();this.$defaults["glow-width"]=t.width,this.$defaults["glow-speed"]=t.speed}catch(t){return console.error(t)}this.$shouldRender=!0}validateColors(){const t={};return[{name:"border-color",value:this["border-color"]},{name:"background-color",value:this["background-color"]},{name:"glow-color",value:this["glow-color"]},{name:"color",value:this.color}].forEach(e=>{const i=e.name;if(e.value.includes(":")){const s=ot(e.value);if(nt(s.light)===void 0||nt(s.dark)===void 0)throw new Error(`Invalid "${i}" value.`);t[i]={light:s.light,dark:s.dark}}else{if(nt(this[i])===void 0)throw new Error(`Invalid "${i}" value.`);t[i]={light:this[i],dark:this[i]}}}),t}validateBorder(){if(ne(this["border-style"])===void 0)throw new Error('Invalid "border-style" value');[{name:"border-radius",value:this["border-radius"]},{name:"border-width",value:this["border-width"]}].forEach(s=>{const o=s.name;if(lt(this[o])===void 0)throw new Error(`Invalid "${o}" value`)});const t={},[e,i="px"]=this["border-radius"].split(/(px|cm|mm|in|pt|pc|em|ex|ch|rem|vw|vh|vmin|vmax|%)/);return t.style=this["border-style"],t.radius={parent:`${Number(e)*1.25}${i}`,child:this["border-radius"]},t.width=this["border-width"],t}validateSize(){return[{name:"width",value:this.width},{name:"height",value:this.height}].forEach(t=>{const e=t.name;if(!["auto","inherit","initial","unset"].includes(this[e])&&lt(this[e])===void 0)throw new Error(`Invalid "${e}" value`)}),{width:this.width,height:this.height}}validateGlow(){if(lt(this["glow-width"])===void 0)throw new Error('Invalid "glow-width" value');if(le(this["glow-speed"])===void 0)throw new Error('Invalid "glow-speed" value');return{width:this["glow-width"],speed:this["glow-speed"]}}render(){const{width:t,["glow-width"]:e,border:i,colors:s,height:o}=this.$defaults;return this.$shouldRender===!0?g`
          <style>
            .p {
              position: relative;
              z-index: 10;
              display: flex;
              align-items: center;
              overflow: hidden;
              width: ${t};
              padding: ${e};
              border-radius: ${i.radius.parent};
              border-width: ${i.width};
              border-style: ${i.style};
              border-color: ${s["border-color"][this.$mode]};
            }

            .p::before {
              content: '';
              position: absolute;
              inset: 0px;
              height: 100%;
              width: 100%;
              animation-name: rotate;
              animation-duration: ${this.$defaults["glow-speed"]};
              animation-timing-function: linear;
              animation-iteration-count: infinite;
              border-radius: 9999px;
              background-image: conic-gradient(
                ${s["glow-color"][this.$mode]} 20deg,
                transparent 120deg
              );
            }

            @keyframes rotate {
              0% {
                transform: rotate(0deg) scale(10);
              }

              100% {
                transform: rotate(-360deg) scale(10);
              }
            }

            .c {
              position: relative;
              z-index: 20;
              width: 100%;
              overflow: hidden;
              height: ${o};
              background-color: ${s["background-color"][this.$mode]};
              color: ${s.color[this.$mode]};
              border-radius: ${i.radius.child};
            }
          </style>

          <div class="p">
            <div class="c">
              <slot></slot>
            </div>
          </div>
        `:""}},w([v()],l.Glow.prototype,"$mode",2),w([v()],l.Glow.prototype,"$defaults",2),w([v()],l.Glow.prototype,"$shouldRender",2),w([c({type:String})],l.Glow.prototype,"border-radius",2),w([c({type:String})],l.Glow.prototype,"border-width",2),w([c({type:String})],l.Glow.prototype,"border-style",2),w([c({type:String})],l.Glow.prototype,"border-color",2),w([c({type:String})],l.Glow.prototype,"width",2),w([c({type:String})],l.Glow.prototype,"height",2),w([c({type:String})],l.Glow.prototype,"background-color",2),w([c({type:String})],l.Glow.prototype,"glow-color",2),w([c({type:String})],l.Glow.prototype,"glow-width",2),w([c({type:String})],l.Glow.prototype,"glow-speed",2),w([c({type:String})],l.Glow.prototype,"color",2),l.Glow=w([q("uk-glow")],l.Glow);var fe=typeof globalThis<"u"?globalThis:typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{};function $e(r){return r&&r.__esModule&&Object.prototype.hasOwnProperty.call(r,"default")?r.default:r}var Lt={exports:{}};(function(r,t){(function(e,i,s){r.exports=s(),r.exports.default=s()})("slugify",fe,function(){var e=JSON.parse(`{"$":"dollar","%":"percent","&":"and","<":"less",">":"greater","|":"or","¢":"cent","£":"pound","¤":"currency","¥":"yen","©":"(c)","ª":"a","®":"(r)","º":"o","À":"A","Á":"A","Â":"A","Ã":"A","Ä":"A","Å":"A","Æ":"AE","Ç":"C","È":"E","É":"E","Ê":"E","Ë":"E","Ì":"I","Í":"I","Î":"I","Ï":"I","Ð":"D","Ñ":"N","Ò":"O","Ó":"O","Ô":"O","Õ":"O","Ö":"O","Ø":"O","Ù":"U","Ú":"U","Û":"U","Ü":"U","Ý":"Y","Þ":"TH","ß":"ss","à":"a","á":"a","â":"a","ã":"a","ä":"a","å":"a","æ":"ae","ç":"c","è":"e","é":"e","ê":"e","ë":"e","ì":"i","í":"i","î":"i","ï":"i","ð":"d","ñ":"n","ò":"o","ó":"o","ô":"o","õ":"o","ö":"o","ø":"o","ù":"u","ú":"u","û":"u","ü":"u","ý":"y","þ":"th","ÿ":"y","Ā":"A","ā":"a","Ă":"A","ă":"a","Ą":"A","ą":"a","Ć":"C","ć":"c","Č":"C","č":"c","Ď":"D","ď":"d","Đ":"DJ","đ":"dj","Ē":"E","ē":"e","Ė":"E","ė":"e","Ę":"e","ę":"e","Ě":"E","ě":"e","Ğ":"G","ğ":"g","Ģ":"G","ģ":"g","Ĩ":"I","ĩ":"i","Ī":"i","ī":"i","Į":"I","į":"i","İ":"I","ı":"i","Ķ":"k","ķ":"k","Ļ":"L","ļ":"l","Ľ":"L","ľ":"l","Ł":"L","ł":"l","Ń":"N","ń":"n","Ņ":"N","ņ":"n","Ň":"N","ň":"n","Ō":"O","ō":"o","Ő":"O","ő":"o","Œ":"OE","œ":"oe","Ŕ":"R","ŕ":"r","Ř":"R","ř":"r","Ś":"S","ś":"s","Ş":"S","ş":"s","Š":"S","š":"s","Ţ":"T","ţ":"t","Ť":"T","ť":"t","Ũ":"U","ũ":"u","Ū":"u","ū":"u","Ů":"U","ů":"u","Ű":"U","ű":"u","Ų":"U","ų":"u","Ŵ":"W","ŵ":"w","Ŷ":"Y","ŷ":"y","Ÿ":"Y","Ź":"Z","ź":"z","Ż":"Z","ż":"z","Ž":"Z","ž":"z","Ə":"E","ƒ":"f","Ơ":"O","ơ":"o","Ư":"U","ư":"u","ǈ":"LJ","ǉ":"lj","ǋ":"NJ","ǌ":"nj","Ș":"S","ș":"s","Ț":"T","ț":"t","ə":"e","˚":"o","Ά":"A","Έ":"E","Ή":"H","Ί":"I","Ό":"O","Ύ":"Y","Ώ":"W","ΐ":"i","Α":"A","Β":"B","Γ":"G","Δ":"D","Ε":"E","Ζ":"Z","Η":"H","Θ":"8","Ι":"I","Κ":"K","Λ":"L","Μ":"M","Ν":"N","Ξ":"3","Ο":"O","Π":"P","Ρ":"R","Σ":"S","Τ":"T","Υ":"Y","Φ":"F","Χ":"X","Ψ":"PS","Ω":"W","Ϊ":"I","Ϋ":"Y","ά":"a","έ":"e","ή":"h","ί":"i","ΰ":"y","α":"a","β":"b","γ":"g","δ":"d","ε":"e","ζ":"z","η":"h","θ":"8","ι":"i","κ":"k","λ":"l","μ":"m","ν":"n","ξ":"3","ο":"o","π":"p","ρ":"r","ς":"s","σ":"s","τ":"t","υ":"y","φ":"f","χ":"x","ψ":"ps","ω":"w","ϊ":"i","ϋ":"y","ό":"o","ύ":"y","ώ":"w","Ё":"Yo","Ђ":"DJ","Є":"Ye","І":"I","Ї":"Yi","Ј":"J","Љ":"LJ","Њ":"NJ","Ћ":"C","Џ":"DZ","А":"A","Б":"B","В":"V","Г":"G","Д":"D","Е":"E","Ж":"Zh","З":"Z","И":"I","Й":"J","К":"K","Л":"L","М":"M","Н":"N","О":"O","П":"P","Р":"R","С":"S","Т":"T","У":"U","Ф":"F","Х":"H","Ц":"C","Ч":"Ch","Ш":"Sh","Щ":"Sh","Ъ":"U","Ы":"Y","Ь":"","Э":"E","Ю":"Yu","Я":"Ya","а":"a","б":"b","в":"v","г":"g","д":"d","е":"e","ж":"zh","з":"z","и":"i","й":"j","к":"k","л":"l","м":"m","н":"n","о":"o","п":"p","р":"r","с":"s","т":"t","у":"u","ф":"f","х":"h","ц":"c","ч":"ch","ш":"sh","щ":"sh","ъ":"u","ы":"y","ь":"","э":"e","ю":"yu","я":"ya","ё":"yo","ђ":"dj","є":"ye","і":"i","ї":"yi","ј":"j","љ":"lj","њ":"nj","ћ":"c","ѝ":"u","џ":"dz","Ґ":"G","ґ":"g","Ғ":"GH","ғ":"gh","Қ":"KH","қ":"kh","Ң":"NG","ң":"ng","Ү":"UE","ү":"ue","Ұ":"U","ұ":"u","Һ":"H","һ":"h","Ә":"AE","ә":"ae","Ө":"OE","ө":"oe","Ա":"A","Բ":"B","Գ":"G","Դ":"D","Ե":"E","Զ":"Z","Է":"E'","Ը":"Y'","Թ":"T'","Ժ":"JH","Ի":"I","Լ":"L","Խ":"X","Ծ":"C'","Կ":"K","Հ":"H","Ձ":"D'","Ղ":"GH","Ճ":"TW","Մ":"M","Յ":"Y","Ն":"N","Շ":"SH","Չ":"CH","Պ":"P","Ջ":"J","Ռ":"R'","Ս":"S","Վ":"V","Տ":"T","Ր":"R","Ց":"C","Փ":"P'","Ք":"Q'","Օ":"O''","Ֆ":"F","և":"EV","ء":"a","آ":"aa","أ":"a","ؤ":"u","إ":"i","ئ":"e","ا":"a","ب":"b","ة":"h","ت":"t","ث":"th","ج":"j","ح":"h","خ":"kh","د":"d","ذ":"th","ر":"r","ز":"z","س":"s","ش":"sh","ص":"s","ض":"dh","ط":"t","ظ":"z","ع":"a","غ":"gh","ف":"f","ق":"q","ك":"k","ل":"l","م":"m","ن":"n","ه":"h","و":"w","ى":"a","ي":"y","ً":"an","ٌ":"on","ٍ":"en","َ":"a","ُ":"u","ِ":"e","ْ":"","٠":"0","١":"1","٢":"2","٣":"3","٤":"4","٥":"5","٦":"6","٧":"7","٨":"8","٩":"9","پ":"p","چ":"ch","ژ":"zh","ک":"k","گ":"g","ی":"y","۰":"0","۱":"1","۲":"2","۳":"3","۴":"4","۵":"5","۶":"6","۷":"7","۸":"8","۹":"9","฿":"baht","ა":"a","ბ":"b","გ":"g","დ":"d","ე":"e","ვ":"v","ზ":"z","თ":"t","ი":"i","კ":"k","ლ":"l","მ":"m","ნ":"n","ო":"o","პ":"p","ჟ":"zh","რ":"r","ს":"s","ტ":"t","უ":"u","ფ":"f","ქ":"k","ღ":"gh","ყ":"q","შ":"sh","ჩ":"ch","ც":"ts","ძ":"dz","წ":"ts","ჭ":"ch","ხ":"kh","ჯ":"j","ჰ":"h","Ṣ":"S","ṣ":"s","Ẁ":"W","ẁ":"w","Ẃ":"W","ẃ":"w","Ẅ":"W","ẅ":"w","ẞ":"SS","Ạ":"A","ạ":"a","Ả":"A","ả":"a","Ấ":"A","ấ":"a","Ầ":"A","ầ":"a","Ẩ":"A","ẩ":"a","Ẫ":"A","ẫ":"a","Ậ":"A","ậ":"a","Ắ":"A","ắ":"a","Ằ":"A","ằ":"a","Ẳ":"A","ẳ":"a","Ẵ":"A","ẵ":"a","Ặ":"A","ặ":"a","Ẹ":"E","ẹ":"e","Ẻ":"E","ẻ":"e","Ẽ":"E","ẽ":"e","Ế":"E","ế":"e","Ề":"E","ề":"e","Ể":"E","ể":"e","Ễ":"E","ễ":"e","Ệ":"E","ệ":"e","Ỉ":"I","ỉ":"i","Ị":"I","ị":"i","Ọ":"O","ọ":"o","Ỏ":"O","ỏ":"o","Ố":"O","ố":"o","Ồ":"O","ồ":"o","Ổ":"O","ổ":"o","Ỗ":"O","ỗ":"o","Ộ":"O","ộ":"o","Ớ":"O","ớ":"o","Ờ":"O","ờ":"o","Ở":"O","ở":"o","Ỡ":"O","ỡ":"o","Ợ":"O","ợ":"o","Ụ":"U","ụ":"u","Ủ":"U","ủ":"u","Ứ":"U","ứ":"u","Ừ":"U","ừ":"u","Ử":"U","ử":"u","Ữ":"U","ữ":"u","Ự":"U","ự":"u","Ỳ":"Y","ỳ":"y","Ỵ":"Y","ỵ":"y","Ỷ":"Y","ỷ":"y","Ỹ":"Y","ỹ":"y","–":"-","‘":"'","’":"'","“":"\\"","”":"\\"","„":"\\"","†":"+","•":"*","…":"...","₠":"ecu","₢":"cruzeiro","₣":"french franc","₤":"lira","₥":"mill","₦":"naira","₧":"peseta","₨":"rupee","₩":"won","₪":"new shequel","₫":"dong","€":"euro","₭":"kip","₮":"tugrik","₯":"drachma","₰":"penny","₱":"peso","₲":"guarani","₳":"austral","₴":"hryvnia","₵":"cedi","₸":"kazakhstani tenge","₹":"indian rupee","₺":"turkish lira","₽":"russian ruble","₿":"bitcoin","℠":"sm","™":"tm","∂":"d","∆":"delta","∑":"sum","∞":"infinity","♥":"love","元":"yuan","円":"yen","﷼":"rial","ﻵ":"laa","ﻷ":"laa","ﻹ":"lai","ﻻ":"la"}`),i=JSON.parse('{"bg":{"Й":"Y","Ц":"Ts","Щ":"Sht","Ъ":"A","Ь":"Y","й":"y","ц":"ts","щ":"sht","ъ":"a","ь":"y"},"de":{"Ä":"AE","ä":"ae","Ö":"OE","ö":"oe","Ü":"UE","ü":"ue","ß":"ss","%":"prozent","&":"und","|":"oder","∑":"summe","∞":"unendlich","♥":"liebe"},"es":{"%":"por ciento","&":"y","<":"menor que",">":"mayor que","|":"o","¢":"centavos","£":"libras","¤":"moneda","₣":"francos","∑":"suma","∞":"infinito","♥":"amor"},"fr":{"%":"pourcent","&":"et","<":"plus petit",">":"plus grand","|":"ou","¢":"centime","£":"livre","¤":"devise","₣":"franc","∑":"somme","∞":"infini","♥":"amour"},"pt":{"%":"porcento","&":"e","<":"menor",">":"maior","|":"ou","¢":"centavo","∑":"soma","£":"libra","∞":"infinito","♥":"amor"},"uk":{"И":"Y","и":"y","Й":"Y","й":"y","Ц":"Ts","ц":"ts","Х":"Kh","х":"kh","Щ":"Shch","щ":"shch","Г":"H","г":"h"},"vi":{"Đ":"D","đ":"d"},"da":{"Ø":"OE","ø":"oe","Å":"AA","å":"aa","%":"procent","&":"og","|":"eller","$":"dollar","<":"mindre end",">":"større end"},"nb":{"&":"og","Å":"AA","Æ":"AE","Ø":"OE","å":"aa","æ":"ae","ø":"oe"},"it":{"&":"e"},"nl":{"&":"en"},"sv":{"&":"och","Å":"AA","Ä":"AE","Ö":"OE","å":"aa","ä":"ae","ö":"oe"}}');function s(o,n){if(typeof o!="string")throw new Error("slugify: string argument expected");n=typeof n=="string"?{replacement:n}:n||{};var h=i[n.locale]||{},a=n.replacement===void 0?"-":n.replacement,p=n.trim===void 0?!0:n.trim,f=o.normalize().split("").reduce(function(u,$){var d=h[$];return d===void 0&&(d=e[$]),d===void 0&&(d=$),d===a&&(d=" "),u+d.replace(n.remove||/[^\w\s$*_+~.()'"!\-:@]+/g,"")},"");return n.strict&&(f=f.replace(/[^A-Za-z0-9\s]/g,"")),p&&(f=f.trim()),f=f.replace(/\s+/g,a),n.lower&&(f=f.toLowerCase()),f}return s.extend=function(o){Object.assign(e,o)},s})})(Lt);var ge=Lt.exports;const me=$e(ge);var ve=Object.defineProperty,ye=Object.getOwnPropertyDescriptor,A=(r,t,e,i)=>{for(var s=i>1?void 0:i?ye(t,e):t,o=r.length-1,n;o>=0;o--)(n=r[o])&&(s=(i?n(t,e,s):n(s))||s);return i&&s&&ve(t,e,s),s};l.InputTag=class extends k{constructor(){super(...arguments),this.disabled=!1,this.error=!1,this.maxlength=20,this.minlength=1,this.name="",this.placeholder="",this.slugify=!1,this["slugify-options"]="",this.state="secondary",this.value="",this.$input="",this.$slugOptions={lower:!0,strict:!0},this.$tags=[]}connectedCallback(){super.connectedCallback(),this.initializeDefaults(),this.removeAttribute("uk-cloak")}createRenderRoot(){return this}initializeDefaults(){if(this.$tags=this.value===""?[]:this.value.split(","),this["slugify-options"]){const t=ot(this["slugify-options"]);"replacement"in t&&(this.$slugOptions.replacement=t.replacement),"remove"in t&&(this.$slugOptions.remove=new RegExp(t.remove,"g")),"lower"in t&&(this.$slugOptions.lower=t.lower==="true"),"strict"in t&&(this.$slugOptions.strict=t.strict==="true"),"locale"in t&&(this.$slugOptions.locale=t.locale),"trim"in t&&(this.$slugOptions.trim=t.trim==="true")}}push(){let t=this.$input;this.slugify&&(t=me(this.$input,this.$slugOptions)),this.$input.length>=this.minlength&&!this.$tags.includes(t)&&(this.$tags.push(t),this.$input=""),this.dispatchEvent(new CustomEvent("uk-input-tag:input",{detail:{value:this.$tags},bubbles:!0,composed:!0}))}render(){return g`
      <div
        class="uk-input-tag ${this.disabled===!0?"opacity-50":""} ${this.error===!0?"uk-form-danger":""}"
      >
        ${this.$tags.map((t,e)=>g`
            <div class="uk-tag ${`uk-tag-${this.state}`}">
              <span
                @click=${()=>{var i;this.disabled===!1&&(this.$input=this.$tags[e],this.$tags=this.$tags.filter((s,o)=>o!==e),(i=this.renderRoot.querySelector("input"))==null||i.focus())}}
              >
                ${t}
              </span>
              <a
                @click="${()=>{this.disabled===!1&&(this.$tags=this.$tags.filter((i,s)=>s!==e))}}"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </a>
            </div>
          `)}

        <input
          .disabled=${this.disabled}
          autocomplete="off"
          type="text"
          placeholder="${this.placeholder}"
          @keydown=${t=>{switch(t.key){case"Backspace":this.$tags.length>0&&this.$input.length===0&&(t.preventDefault(),this.$input=this.$tags.slice(-1)[0],this.$tags.pop());break;case",":t.preventDefault(),this.push();break;case"Enter":t.preventDefault(),this.push();break}}}
          @input=${t=>{const e=t.target;this.$input=e.value}}
          .maxLength=${this.maxlength}
          .value=${this.$input}
        />

        ${this.$tags.map(t=>g`
            <input name="${this.name}[]" type="hidden" value="${t}" />
          `)}
      </div>
    `}},A([c({type:Boolean})],l.InputTag.prototype,"disabled",2),A([c({type:Boolean})],l.InputTag.prototype,"error",2),A([c({type:Number})],l.InputTag.prototype,"maxlength",2),A([c({type:Number})],l.InputTag.prototype,"minlength",2),A([c({type:String})],l.InputTag.prototype,"name",2),A([c({type:String})],l.InputTag.prototype,"placeholder",2),A([c({type:Boolean})],l.InputTag.prototype,"slugify",2),A([c({type:String})],l.InputTag.prototype,"slugify-options",2),A([c({type:String})],l.InputTag.prototype,"state",2),A([c({type:String})],l.InputTag.prototype,"value",2),A([v()],l.InputTag.prototype,"$input",2),A([v()],l.InputTag.prototype,"$slugOptions",2),A([v()],l.InputTag.prototype,"$tags",2),l.InputTag=A([q("uk-input-tag")],l.InputTag);var be=Object.defineProperty,we=Object.getOwnPropertyDescriptor,E=(r,t,e,i)=>{for(var s=i>1?void 0:i?we(t,e):t,o=r.length-1,n;o>=0;o--)(n=r[o])&&(s=(i?n(t,e,s):n(s))||s);return i&&s&&be(t,e,s),s};l.InputPin=class extends k{constructor(){super(...arguments),this.autofocus=!1,this.disabled=!1,this.error=!1,this.length=6,this.name="",this.separated=!1,this.$value=""}connectedCallback(){super.connectedCallback(),this.removeAttribute("uk-cloak")}createRenderRoot(){return this}updated(t){t.has("$value")&&t.get("$value")!==this.$value&&this.dispatchEvent(new CustomEvent("uk-input-pin:input",{detail:{value:this.$value},bubbles:!0,composed:!0}))}render(){return g`
      <div
        class="uk-input-pin ${this.separated===!0?"uk-input-pin-separated":""} ${this.disabled===!0?"uk-disabled":""} ${this.error===!0?"uk-form-danger":""}"
      >
        ${Array(this.length).fill("").map((t,e)=>g`<input
                type="text"
                maxlength="1"
                placeholder="○"
                .autofocus="${!!(this.autofocus&&e===0)}"
                .disabled=${this.disabled}
                @keydown="${i=>{const s=this.renderRoot.querySelectorAll('input[type="text"]');switch(i.key){case"Backspace":const o=i.target;this.$focus!==void 0&&o.value.length===0&&this.$focus>0&&(i.preventDefault(),s[this.$focus-1].focus());break}}}"
                @input="${i=>{const s=this.renderRoot.querySelectorAll('input[type="text"]');let o="";s.forEach(h=>{o+=h.value});const n=i.target;n.value.length===1&&(e<this.length-1&&s[e+1].focus(),e===this.length-1&&n.blur()),this.$value=o}}"
                @focus="${()=>this.$focus=e}"
                @blur="${()=>this.$focus=void 0}"
              />`)}
      </div>
      <input type="hidden" name="${this.name}" .value=${this.$value} />
    `}},E([c({type:Boolean})],l.InputPin.prototype,"autofocus",2),E([c({type:Boolean})],l.InputPin.prototype,"disabled",2),E([c({type:Boolean})],l.InputPin.prototype,"error",2),E([c({type:Number})],l.InputPin.prototype,"length",2),E([c({type:String})],l.InputPin.prototype,"name",2),E([c({type:Boolean})],l.InputPin.prototype,"separated",2),E([v()],l.InputPin.prototype,"$focus",2),E([v()],l.InputPin.prototype,"$value",2),l.InputPin=E([q("uk-input-pin")],l.InputPin);class at extends Rt{constructor(t){if(super(t),this.it=m,t.type!==It.CHILD)throw Error(this.constructor.directiveName+"() can only be used in child bindings")}render(t){if(t===m||t==null)return this._t=void 0,this.it=t;if(t===C)return t;if(typeof t!="string")throw Error(this.constructor.directiveName+"() called with a non-string value");if(t===this.it)return this._t;this.it=t;const e=[t];return e.raw=e,this._t={_$litType$:this.constructor.resultType,strings:e,values:[]}}}at.directiveName="unsafeHTML",at.resultType=1;const Ae=Ut(at);var _e=Object.defineProperty,ke=Object.getOwnPropertyDescriptor,P=(r,t,e,i)=>{for(var s=i>1?void 0:i?ke(t,e):t,o=r.length-1,n;o>=0;o--)(n=r[o])&&(s=(i?n(t,e,s):n(s))||s);return i&&s&&_e(t,e,s),s};return l.Command=class extends k{constructor(){super(),this.placeholder="Search",this.$items={},this.$filteredItems=this.$items,this.$term="",this.$focused=-1,this.toggle=ae()}get flattenedItems(){return Object.values(this.$filteredItems).flat()}connectedCallback(){super.connectedCallback(),Array.from(this.children).map(t=>{var e,i;if(t.nodeName==="A"){const s=t.hasAttribute("data-group")?t.getAttribute("data-group"):"__";this.$items[s]===void 0&&(this.$items[s]=[]);let o=[(e=t.textContent)==null?void 0:e.toLocaleLowerCase().trim()];if(t.hasAttribute("data-keywords")){const n=(i=t.getAttribute("data-keywords"))==null?void 0:i.split(",").map(h=>h.trim()).filter(h=>h!=="");o=[...o,...n]}this.$items[s].push({disabled:t.hasAttribute("href")===!1||t.getAttribute("href")==="",element:t.outerHTML,keywords:o})}}),this.key!==void 0&&document.addEventListener("keydown",this.onKeydown.bind(this)),this.innerHTML="",this.removeAttribute("uk-cloak")}disconnectedCallback(){this.key!==void 0&&document.removeEventListener("keydown",this.onKeydown)}createRenderRoot(){return this}updated(t){if(t.has("$focused")){const e=document.getElementById(this.toggle);if(e){const i=e.querySelector("ul");if(i){const s=i.querySelectorAll("li");s.forEach(n=>{n.classList.remove("uk-active")});const o=Array.from(s).filter(n=>!n.classList.contains("uk-nav-header"));if(this.$focused>=0&&this.$focused<o.length){const n=o[this.$focused];n.classList.add("uk-active");const h={ul:i.getBoundingClientRect(),li:n.getBoundingClientRect()},a=n.offsetTop-i.offsetTop-h.ul.height/2+h.li.height/2;i.scrollTo({top:a,behavior:"smooth"})}}}}t.has("$term")&&this.updateComplete.then(()=>{if(this.$focused=-1,this.$term===""){this.$filteredItems=this.$items;return}this.$filteredItems=Object.fromEntries(Object.entries(this.$items).map(([e,i])=>[e,i.filter(s=>s.keywords.some(o=>o.toLowerCase().includes(this.$term.toLowerCase())))]))})}navigate(t){const e=o=>o.disabled!==!0;let i=this.$focused;const s=t==="up"?-1:1;do if(i+=s,i<0){for(i=this.flattenedItems.length-1;i>=0&&!e(this.flattenedItems[i]);)i--;break}else if(i>=this.flattenedItems.length){for(i=0;i<this.flattenedItems.length&&!e(this.flattenedItems[i]);)i++;break}while(!e(this.flattenedItems[i]));return i}go(){const t=document.getElementById(this.toggle);if(t){const e=t.querySelector("ul");if(e){const i=e.querySelectorAll("li"),s=Array.from(i).filter(o=>!o.classList.contains("uk-nav-header"));s[this.$focused]&&s[this.$focused].querySelector("a").click()}}}render(){return g`
      <div id=${this.toggle} class="uk-modal uk-flex-top" uk-modal>
        <div class="uk-margin-auto-vertical uk-modal-dialog">
          <div class="uk-inline uk-width-1-1">
            <span class="uk-form-icon uk-form-icon-flip uk-text-muted">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.3-4.3" />
              </svg>
            </span>
            <input
              autofocus
              class="uk-input uk-form-blank"
              placeholder="${this.placeholder}"
              type="text"
              @keydown=${t=>{switch(t.key){case"ArrowDown":t.preventDefault(),this.$focused=this.navigate("down");break;case"ArrowUp":t.preventDefault(),this.$focused=this.navigate("up");break;case"Enter":t.preventDefault(),this.go();break}}}
              @input=${t=>{const e=t.target;this.$term=e.value}}
            />
          </div>
          <hr class="uk-hr" />
          <ul class="uk-height-medium uk-overflow-auto uk-nav uk-nav-secondary">
            ${rt(Object.entries(this.$filteredItems),([t])=>t,([t,e])=>e.length?g`
                      ${t!=="__"?g`<li class="uk-nav-header">${t}</li>`:""}
                      ${rt(e,(i,s)=>s,i=>g`
                          <li
                            class="${i.disabled===!0?"uk-disabled opacity-50":""}"
                          >
                            ${Ae(i.element)}
                          </li>
                        `)}
                    `:"")}
          </ul>
        </div>
      </div>
    `}onKeydown(t){t.ctrlKey&&t.key===this.key&&(t.preventDefault(),window.UIkit.modal(`#${this.toggle}`).toggle())}},P([c({type:String})],l.Command.prototype,"key",2),P([c({type:String})],l.Command.prototype,"placeholder",2),P([c({type:String})],l.Command.prototype,"toggle",2),P([v()],l.Command.prototype,"$items",2),P([v()],l.Command.prototype,"$filteredItems",2),P([v()],l.Command.prototype,"$term",2),P([v()],l.Command.prototype,"$focused",2),l.Command=P([q("uk-command")],l.Command),Object.defineProperty(l,Symbol.toStringTag,{value:"Module"}),l}({});
/**
 * @license
 * Copyright 2019 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
/**
 * @license
 * Copyright 2020 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */