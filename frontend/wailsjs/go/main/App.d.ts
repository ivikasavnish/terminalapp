// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';

export function AddCommandToHistory(arg1:string,arg2:string):Promise<void>;

export function ConnectSSHWithHostKeyCheck(arg1:string):Promise<main.ConnectionResult>;

export function CreateSynonym(arg1:string):Promise<string>;

export function DeleteCustomProfile(arg1:string):Promise<void>;

export function DisconnectSSH(arg1:string):Promise<void>;

export function GetActiveConnections():Promise<Array<string>>;

export function GetBaseProfile():Promise<main.SSHConfig>;

export function GetCommandHistory(arg1:string):Promise<Array<string>>;

export function LoadProfiles():Promise<Array<main.SSHConfig>>;

export function SaveCustomProfile(arg1:main.CustomProfile):Promise<void>;
