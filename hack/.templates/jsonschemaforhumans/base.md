{{ schema.keywords.get("title").literal | default("Jackal Package Schema") | md_heading(0) }}
{% set contentBase %}
{% with schema=schema, skip_headers=False %}
    {% include "content.md" %}
{% endwith %}
{% endset %}

{{ md_get_toc() }}

{{ contentBase }}

----------------------------------------------------------------------------------------------------------------------------
{% if config.with_footer -%}
Generated from [jackal.schema.json](https://github.com/Racer159/jackal/blob/main/jackal.schema.json){% if config.footer_show_time %} on {{ get_local_time() }}{% endif %}
{%- endif -%}

{{ "\n" }}
