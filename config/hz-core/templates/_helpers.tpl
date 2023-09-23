{{/*
Expand the name of the chart.
*/}}
{{- define "hz-core.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "hz-core.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}


{{- define "hz-core.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "hz-core.labels" -}}
helm.sh/chart: {{ include "hz-core.chart" . }}
{{ include "hz-core.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{- define "hz-core.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hz-core.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "hz-core.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "hz-core.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "getOrDefaultPass" -}}
{{- if not .Values.adminPassword -}}
{{- printf "$2a$10$zcHepmzfKPoxCVCYZr5K7ORPZZ/ySe9p/7IUb/8u./xHrnSX2LOCO" -}}
{{- else -}}
{{- printf "%s" .Values.adminPassword -}}
{{- end -}}
{{- end -}}
