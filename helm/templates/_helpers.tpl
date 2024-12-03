{{/*
    Helper template to generate a fullname.
    */}}
    {{- define "rekor.fullname" -}}
    {{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
    {{- end -}}
    